// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package bucket

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcapitypes "github.com/aws-controllers-k8s/s3-controller/apis/v1alpha1"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/s3"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	s3controltypes "github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/aws/smithy-go"
)

// IsDirectoryBucketName checks if a bucket name follows the directory bucket naming pattern.
// Directory bucket names must end with "--x-s3" suffix.
func IsDirectoryBucketName(bucketName string) bool {
	return strings.HasSuffix(bucketName, "--x-s3")
}

// bucketARN returns the ARN of the S3 bucket with the given name.
func bucketARN(bucketName string) string {
	// TODO(a-hilaly): I know there could be other partitions, but I'm
	// not sure how to determine at this level of abstraction. Probably
	// something the SDK/runtime should handle. For now, we'll just use
	// the `aws` partition.
	//
	// NOTE(a-hilaly): Other parts of ACK also use this default partition
	// e.g the generated function `ARNFromName` also uses `aws` as the
	// default partition.
	return fmt.Sprintf("arn:aws:s3:::%s", bucketName)
}

// directoryBucketARN returns the fully-qualified ARN for a directory bucket.
// S3 Control API requires region and account ID in the ARN for directory buckets.
func (rm *resourceManager) directoryBucketARN(bucketName string) string {
	return fmt.Sprintf("arn:aws:s3express:%s:%s:bucket/%s",
		string(rm.awsRegion),
		string(rm.awsAccountID),
		bucketName)
}

var (
	DefaultAccelerationStatus     = svcsdktypes.BucketAccelerateStatusSuspended
	DefaultRequestPayer           = svcsdktypes.PayerBucketOwner
	DefaultVersioningStatus       = svcsdktypes.BucketVersioningStatusSuspended
	DefaultACL                    = svcsdktypes.BucketCannedACLPrivate
	DefaultPublicBlockAccessValue = false
	DefaultPublicBlockAccess      = svcapitypes.PublicAccessBlockConfiguration{
		BlockPublicACLs:       &DefaultPublicBlockAccessValue,
		BlockPublicPolicy:     &DefaultPublicBlockAccessValue,
		IgnorePublicACLs:      &DefaultPublicBlockAccessValue,
		RestrictPublicBuckets: &DefaultPublicBlockAccessValue,
	}
	CannedACLJoinDelimiter = "|"
)

// ConfigurationAction stores the possible actions that can be performed on
// any of the elements of a configuration list
type ConfigurationAction int

const (
	ConfigurationActionNone ConfigurationAction = iota
	ConfigurationActionPut
	ConfigurationActionDelete
	ConfigurationActionUpdate
)

const ErrSyncingPutProperty = "Error syncing property '%s'"

// validateDirectoryBucketSpec validates that no unsupported fields are set for directory buckets.
// Returns a terminal error if unsupported fields are specified.
func validateDirectoryBucketSpec(ko *svcapitypes.Bucket) error {
	if ko.Spec.Name == nil || !IsDirectoryBucketName(*ko.Spec.Name) {
		return nil
	}

	unsupportedChecks := []struct {
		name  string
		isSet func() bool
	}{
		{"Accelerate", func() bool { return ko.Spec.Accelerate != nil }},
		{"Analytics", func() bool { return len(ko.Spec.Analytics) > 0 }},
		{"ACL", func() bool { return ko.Spec.ACL != nil }},
		{"GrantFullControl", func() bool { return ko.Spec.GrantFullControl != nil }},
		{"GrantRead", func() bool { return ko.Spec.GrantRead != nil }},
		{"GrantReadACP", func() bool { return ko.Spec.GrantReadACP != nil }},
		{"GrantWrite", func() bool { return ko.Spec.GrantWrite != nil }},
		{"GrantWriteACP", func() bool { return ko.Spec.GrantWriteACP != nil }},
		{"CORS", func() bool { return ko.Spec.CORS != nil }},
		{"IntelligentTiering", func() bool { return len(ko.Spec.IntelligentTiering) > 0 }},
		{"Inventory", func() bool { return len(ko.Spec.Inventory) > 0 }},
		{"Logging", func() bool { return ko.Spec.Logging != nil }},
		{"Metrics", func() bool { return len(ko.Spec.Metrics) > 0 }},
		{"Notification", func() bool { return ko.Spec.Notification != nil }},
		{"OwnershipControls", func() bool { return ko.Spec.OwnershipControls != nil }},
		{"PublicAccessBlock", func() bool { return ko.Spec.PublicAccessBlock != nil }},
		{"Replication", func() bool { return ko.Spec.Replication != nil }},
		{"RequestPayment", func() bool { return ko.Spec.RequestPayment != nil }},
		{"Versioning", func() bool { return ko.Spec.Versioning != nil }},
		{"Website", func() bool { return ko.Spec.Website != nil }},
	}

	var unsupportedFields []string
	for _, check := range unsupportedChecks {
		if check.isSet() {
			unsupportedFields = append(unsupportedFields, check.name)
		}
	}

	if len(unsupportedFields) > 0 {
		return ackerr.NewTerminalError(fmt.Errorf(
			"directory buckets do not support the following fields: %s. "+
				"See https://docs.aws.amazon.com/AmazonS3/latest/userguide/s3-express-differences.html",
			strings.Join(unsupportedFields, ", "),
		))
	}

	return nil
}

// customFindBucket is a custom implementation of sdkFind that handles both
// general-purpose and directory buckets by using the appropriate List API.
func (rm *resourceManager) customFindBucket(
	ctx context.Context,
	r *resource,
) (latest *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.customFindBucket")
	defer func() {
		exit(err)
	}()

	if r.ko.Spec.Name == nil {
		return nil, ackerr.NotFound
	}

	bucketName := *r.ko.Spec.Name
	ko := r.ko.DeepCopy()

	if IsDirectoryBucketName(bucketName) {
		// Use ListDirectoryBuckets for directory buckets
		input := &svcsdk.ListDirectoryBucketsInput{}
		found := false

		paginator := svcsdk.NewListDirectoryBucketsPaginator(rm.sdkapi, input)
		for paginator.HasMorePages() {
			resp, err := paginator.NextPage(ctx)
			rm.metrics.RecordAPICall("READ_MANY", "ListDirectoryBuckets", err)
			if err != nil {
				var awsErr smithy.APIError
				if errors.As(err, &awsErr) && awsErr.ErrorCode() == "NoSuchBucket" {
					return nil, ackerr.NotFound
				}
				return nil, err
			}
			for _, bucket := range resp.Buckets {
				if bucket.Name != nil && *bucket.Name == bucketName {
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			return nil, ackerr.NotFound
		}
	} else {
		// Use ListBuckets for general-purpose buckets
		input := &svcsdk.ListBucketsInput{}
		resp, err := rm.sdkapi.ListBuckets(ctx, input)
		rm.metrics.RecordAPICall("READ_MANY", "ListBuckets", err)
		if err != nil {
			var awsErr smithy.APIError
			if errors.As(err, &awsErr) && awsErr.ErrorCode() == "NoSuchBucket" {
				return nil, ackerr.NotFound
			}
			return nil, err
		}

		found := false
		for _, bucket := range resp.Buckets {
			if bucket.Name != nil && *bucket.Name == bucketName {
				found = true
				break
			}
		}

		if !found {
			return nil, ackerr.NotFound
		}
	}

	rm.setStatusDefaults(ko)
	if err := rm.addPutFieldsToSpec(ctx, r, ko); err != nil {
		return nil, err
	}

	// Set bucket ARN in the output
	var arnStr ackv1alpha1.AWSResourceName

	// Directory buckets use fully-qualified ARN for S3 Control API compatibility
	if IsDirectoryBucketName(*ko.Spec.Name) {
		arnStr = ackv1alpha1.AWSResourceName(rm.directoryBucketARN(*ko.Spec.Name))
	} else {
		arnStr = ackv1alpha1.AWSResourceName(bucketARN(*ko.Spec.Name))
	}
	ko.Status.ACKResourceMetadata.ARN = &arnStr

	return &resource{ko}, nil
}

// customUpdateBucket patches each of the resource properties in the backend AWS
// service API and returns a new resource with updated fields.
func (rm *resourceManager) customUpdateBucket(
	ctx context.Context,
	desired *resource,
	latest *resource,
	delta *ackcompare.Delta,
) (updated *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.customUpdateBucket")
	defer exit(err)

	// Validate directory bucket spec before updating
	if err := validateDirectoryBucketSpec(desired.ko); err != nil {
		return nil, err
	}

	isDirectoryBucket := desired.ko.Spec.Name != nil && IsDirectoryBucketName(*desired.ko.Spec.Name)

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := desired.ko.DeepCopy()
	ko.Status = *latest.ko.Status.DeepCopy()

	// Skip unsupported operations for directory buckets
	if !isDirectoryBucket && delta.DifferentAt("Spec.Accelerate") {
		if err := rm.syncAccelerate(ctx, desired); err != nil {
			return nil, errors.Wrapf(err, ErrSyncingPutProperty, "Accelerate")
		}
	}
	if !isDirectoryBucket && delta.DifferentAt("Spec.Analytics") {
		if err := rm.syncAnalytics(ctx, desired, latest); err != nil {
			return nil, errors.Wrapf(err, ErrSyncingPutProperty, "Analytics")
		}
	}
	if !isDirectoryBucket && (delta.DifferentAt("Spec.ACL") ||
		delta.DifferentAt("Spec.GrantFullControl") ||
		delta.DifferentAt("Spec.GrantRead") ||
		delta.DifferentAt("Spec.GrantReadACP") ||
		delta.DifferentAt("Spec.GrantWrite") ||
		delta.DifferentAt("Spec.GrantWriteACP")) {
		if err := rm.syncACL(ctx, desired); err != nil {
			return nil, errors.Wrapf(err, ErrSyncingPutProperty, "ACLs or Grant Headers")
		}
	}
	if !isDirectoryBucket && delta.DifferentAt("Spec.CORS") {
		if err := rm.syncCORS(ctx, desired); err != nil {
			return nil, errors.Wrapf(err, ErrSyncingPutProperty, "CORS")
		}
	}
	// Encryption is supported for both bucket types
	if delta.DifferentAt("Spec.Encryption") {
		if err := rm.syncEncryption(ctx, desired); err != nil {
			return nil, errors.Wrapf(err, ErrSyncingPutProperty, "Encryption")
		}
	}
	if !isDirectoryBucket && delta.DifferentAt("Spec.IntelligentTiering") {
		if err := rm.syncIntelligentTiering(ctx, desired, latest); err != nil {
			return nil, errors.Wrapf(err, ErrSyncingPutProperty, "IntelligentTiering")
		}
	}
	if !isDirectoryBucket && delta.DifferentAt("Spec.Inventory") {
		if err := rm.syncInventory(ctx, desired, latest); err != nil {
			return nil, errors.Wrapf(err, ErrSyncingPutProperty, "Inventory")
		}
	}
	// Lifecycle is supported for both bucket types
	if delta.DifferentAt("Spec.Lifecycle") {
		if err := rm.syncLifecycle(ctx, desired); err != nil {
			return nil, errors.Wrapf(err, ErrSyncingPutProperty, "Lifecycle")
		}
	}
	if !isDirectoryBucket && delta.DifferentAt("Spec.Logging") {
		if err := rm.syncLogging(ctx, desired); err != nil {
			return nil, errors.Wrapf(err, ErrSyncingPutProperty, "Logging")
		}
	}
	if !isDirectoryBucket && delta.DifferentAt("Spec.Metrics") {
		if err := rm.syncMetrics(ctx, desired, latest); err != nil {
			return nil, errors.Wrapf(err, ErrSyncingPutProperty, "Metrics")
		}
	}
	if !isDirectoryBucket && delta.DifferentAt("Spec.Notification") {
		if err := rm.syncNotification(ctx, desired); err != nil {
			return nil, errors.Wrapf(err, ErrSyncingPutProperty, "Notification")
		}
	}
	if !isDirectoryBucket && delta.DifferentAt("Spec.OwnershipControls") {
		if err := rm.syncOwnershipControls(ctx, desired); err != nil {
			return nil, errors.Wrapf(err, ErrSyncingPutProperty, "OwnershipControls")
		}
	}
	// PublicAccessBlock may need to be set in order to use Policy, so sync it
	// first (not supported for directory buckets)
	if !isDirectoryBucket && delta.DifferentAt("Spec.PublicAccessBlock") {
		if err := rm.syncPublicAccessBlock(ctx, desired); err != nil {
			return nil, errors.Wrapf(err, ErrSyncingPutProperty, "PublicAccessBlock")
		}
	}
	// Policy is supported for both bucket types
	if delta.DifferentAt("Spec.Policy") {
		if err := rm.syncPolicy(ctx, desired, isDirectoryBucket); err != nil {
			return nil, errors.Wrapf(err, ErrSyncingPutProperty, "Policy")
		}
	}
	if !isDirectoryBucket && delta.DifferentAt("Spec.RequestPayment") {
		if err := rm.syncRequestPayment(ctx, desired); err != nil {
			return nil, errors.Wrapf(err, ErrSyncingPutProperty, "RequestPayment")
		}
	}
	// Tagging is supported for both bucket types
	if delta.DifferentAt("Spec.Tagging") {
		if err := rm.syncTagging(ctx, desired); err != nil {
			return nil, errors.Wrapf(err, ErrSyncingPutProperty, "Tagging")
		}
	}
	if !isDirectoryBucket && delta.DifferentAt("Spec.Website") {
		if err := rm.syncWebsite(ctx, desired); err != nil {
			return nil, errors.Wrapf(err, ErrSyncingPutProperty, "Website")
		}
	}

	// Replication requires versioning be enabled. We check that if we are
	// disabling versioning, that we disable replication first. If we are
	// enabling replication, that we enable versioning first.
	// (Not supported for directory buckets)
	if !isDirectoryBucket && (delta.DifferentAt("Spec.Replication") || delta.DifferentAt("Spec.Versioning")) {
		if desired.ko.Spec.Replication == nil || desired.ko.Spec.Replication.Rules == nil {
			if err := rm.syncReplication(ctx, desired); err != nil {
				return nil, errors.Wrapf(err, ErrSyncingPutProperty, "Replication")
			}
			if err := rm.syncVersioning(ctx, desired); err != nil {
				return nil, errors.Wrapf(err, ErrSyncingPutProperty, "Versioning")
			}
		} else {
			if err := rm.syncVersioning(ctx, desired); err != nil {
				return nil, errors.Wrapf(err, ErrSyncingPutProperty, "Versioning")
			}
			if err := rm.syncReplication(ctx, desired); err != nil {
				return nil, errors.Wrapf(err, ErrSyncingPutProperty, "Replication")
			}
		}
	}

	return &resource{ko}, nil
}

// addPutFieldsToSpec will describe each of the Put* fields and add their
// returned values to the Bucket spec.
func (rm *resourceManager) addPutFieldsToSpec(
	ctx context.Context,
	r *resource,
	ko *svcapitypes.Bucket,
) (err error) {
	isDirectoryBucket := r.ko.Spec.Name != nil && IsDirectoryBucketName(*r.ko.Spec.Name)

	// Skip unsupported API calls for directory buckets
	if !isDirectoryBucket {
		getAccelerateResponse, err := rm.sdkapi.GetBucketAccelerateConfiguration(ctx, rm.newGetBucketAcceleratePayload(r))
		if err != nil {
			// This method is not supported in every region, ignore any errors if
			// we attempt to describe this property in a region in which it's not
			// supported.
			if awsErr, ok := ackerr.AWSError(err); !ok || (awsErr.ErrorCode() != "MethodNotAllowed" && awsErr.ErrorCode() != "UnsupportedArgument") {
				return err
			}
		}
		if getAccelerateResponse == nil || getAccelerateResponse.Status == "" {
			ko.Spec.Accelerate = nil
		} else {
			ko.Spec.Accelerate = rm.setResourceAccelerate(r, getAccelerateResponse)
		}

		listAnalyticsResponse, err := rm.sdkapi.ListBucketAnalyticsConfigurations(ctx, rm.newListBucketAnalyticsPayload(r))
		if err != nil {
			return err
		}
		if listAnalyticsResponse != nil && len(listAnalyticsResponse.AnalyticsConfigurationList) > 0 {
			ko.Spec.Analytics = make([]*svcapitypes.AnalyticsConfiguration, len(listAnalyticsResponse.AnalyticsConfigurationList))
			for i, analyticsConfiguration := range listAnalyticsResponse.AnalyticsConfigurationList {
				ko.Spec.Analytics[i] = rm.setResourceAnalyticsConfiguration(r, analyticsConfiguration)
			}
		} else {
			ko.Spec.Analytics = nil
		}

		getACLResponse, err := rm.sdkapi.GetBucketAcl(ctx, rm.newGetBucketACLPayload(r))
		if err != nil {
			return err
		}
		rm.setResourceACL(ko, getACLResponse)

		getCORSResponse, err := rm.sdkapi.GetBucketCors(ctx, rm.newGetBucketCORSPayload(r))
		if err != nil {
			if awsErr, ok := ackerr.AWSError(err); !ok || awsErr.ErrorCode() != "NoSuchCORSConfiguration" {
				return err
			}
		}
		if getCORSResponse != nil {
			ko.Spec.CORS = rm.setResourceCORS(r, getCORSResponse)
		} else {
			ko.Spec.CORS = nil
		}
	}

	// Encryption is supported for both bucket types
	getEncryptionResponse, err := rm.sdkapi.GetBucketEncryption(ctx, rm.newGetBucketEncryptionPayload(r))
	if err != nil {
		if awsErr, ok := ackerr.AWSError(err); !ok || awsErr.ErrorCode() != "ServerSideEncryptionConfigurationNotFoundError" {
			return err
		}
	}
	if getEncryptionResponse.ServerSideEncryptionConfiguration.Rules != nil {
		ko.Spec.Encryption = rm.setResourceEncryption(r, getEncryptionResponse)
	} else {
		ko.Spec.Encryption = nil
	}

	if !isDirectoryBucket {
		listIntelligentTieringResponse, err := rm.sdkapi.ListBucketIntelligentTieringConfigurations(ctx, rm.newListBucketIntelligentTieringPayload(r))
		if err != nil {
			return err
		}
		if len(listIntelligentTieringResponse.IntelligentTieringConfigurationList) > 0 {
			ko.Spec.IntelligentTiering = make([]*svcapitypes.IntelligentTieringConfiguration, len(listIntelligentTieringResponse.IntelligentTieringConfigurationList))
			for i, intelligentTieringConfiguration := range listIntelligentTieringResponse.IntelligentTieringConfigurationList {
				ko.Spec.IntelligentTiering[i] = rm.setResourceIntelligentTieringConfiguration(r, intelligentTieringConfiguration)
			}
		} else {
			ko.Spec.IntelligentTiering = nil
		}

		listInventoryResponse, err := rm.sdkapi.ListBucketInventoryConfigurations(ctx, rm.newListBucketInventoryPayload(r))
		if err != nil {
			return err
		}

		ko.Spec.Inventory = make([]*svcapitypes.InventoryConfiguration, len(listInventoryResponse.InventoryConfigurationList))
		for i, inventoryConfiguration := range listInventoryResponse.InventoryConfigurationList {
			ko.Spec.Inventory[i] = rm.setResourceInventoryConfiguration(r, inventoryConfiguration)
		}
	}

	// Lifecycle is supported for both bucket types
	getLifecycleResponse, err := rm.sdkapi.GetBucketLifecycleConfiguration(ctx, rm.newGetBucketLifecyclePayload(r))
	if err != nil {
		if awsErr, ok := ackerr.AWSError(err); !ok || awsErr.ErrorCode() != "NoSuchLifecycleConfiguration" {
			return err
		}
	}
	if getLifecycleResponse != nil {
		ko.Spec.Lifecycle = rm.setResourceLifecycle(r, getLifecycleResponse)
	} else {
		ko.Spec.Lifecycle = nil
	}

	if !isDirectoryBucket {
		getLoggingResponse, err := rm.sdkapi.GetBucketLogging(ctx, rm.newGetBucketLoggingPayload(r))
		if err != nil {
			return err
		}
		if getLoggingResponse.LoggingEnabled != nil {
			ko.Spec.Logging = rm.setResourceLogging(r, getLoggingResponse)
		} else {
			ko.Spec.Logging = nil
		}

		listMetricsResponse, err := rm.sdkapi.ListBucketMetricsConfigurations(ctx, rm.newListBucketMetricsPayload(r))
		if err != nil {
			return err
		}
		if len(listMetricsResponse.MetricsConfigurationList) > 0 {
			ko.Spec.Metrics = make([]*svcapitypes.MetricsConfiguration, len(listMetricsResponse.MetricsConfigurationList))
			for i, metricsConfiguration := range listMetricsResponse.MetricsConfigurationList {
				ko.Spec.Metrics[i] = rm.setResourceMetricsConfiguration(r, &metricsConfiguration)
			}
		} else {
			ko.Spec.Metrics = nil
		}

		getNotificationResponse, err := rm.sdkapi.GetBucketNotificationConfiguration(ctx, rm.newGetBucketNotificationPayload(r))
		if err != nil {
			return err
		}
		if getNotificationResponse.LambdaFunctionConfigurations != nil ||
			getNotificationResponse.QueueConfigurations != nil ||
			getNotificationResponse.TopicConfigurations != nil {

			ko.Spec.Notification = rm.setResourceNotification(r, getNotificationResponse)
		} else {
			ko.Spec.Notification = nil
		}

		getOwnershipControlsResponse, err := rm.sdkapi.GetBucketOwnershipControls(ctx, rm.newGetBucketOwnershipControlsPayload(r))
		if err != nil {
			if awsErr, ok := ackerr.AWSError(err); !ok || awsErr.ErrorCode() != "OwnershipControlsNotFoundError" {
				return err
			}
		}
		if getOwnershipControlsResponse != nil {
			ko.Spec.OwnershipControls = rm.setResourceOwnershipControls(r, getOwnershipControlsResponse)
		} else {
			ko.Spec.OwnershipControls = nil
		}
	}

	// Policy is supported for both bucket types
	getPolicyResponse, err := rm.sdkapi.GetBucketPolicy(ctx, rm.newGetBucketPolicyPayload(r))
	if err != nil {
		if awsErr, ok := ackerr.AWSError(err); !ok || awsErr.ErrorCode() != "NoSuchBucketPolicy" {
			return err
		}
	}
	if getPolicyResponse != nil {
		ko.Spec.Policy = getPolicyResponse.Policy
	} else {
		ko.Spec.Policy = nil
	}

	if !isDirectoryBucket {
		getPublicAccessBlockResponse, err := rm.sdkapi.GetPublicAccessBlock(ctx, rm.newGetPublicAccessBlockPayload(r))
		if err != nil {
			if awsErr, ok := ackerr.AWSError(err); !ok || awsErr.ErrorCode() != "NoSuchPublicAccessBlockConfiguration" {
				return err
			}
		}
		if getPublicAccessBlockResponse != nil {
			ko.Spec.PublicAccessBlock = rm.setResourcePublicAccessBlock(r, getPublicAccessBlockResponse)
		} else {
			ko.Spec.PublicAccessBlock = nil
		}

		getReplicationResponse, err := rm.sdkapi.GetBucketReplication(ctx, rm.newGetBucketReplicationPayload(r))
		if err != nil {
			if awsErr, ok := ackerr.AWSError(err); !ok || awsErr.ErrorCode() != "ReplicationConfigurationNotFoundError" {
				return err
			}
		}
		if getReplicationResponse != nil {
			ko.Spec.Replication = rm.setResourceReplication(r, getReplicationResponse)
		} else {
			ko.Spec.Replication = nil
		}

		getRequestPaymentResponse, err := rm.sdkapi.GetBucketRequestPayment(ctx, rm.newGetBucketRequestPaymentPayload(r))
		if err != nil {
			return nil
		}
		if getRequestPaymentResponse.Payer != "" {
			ko.Spec.RequestPayment = rm.setResourceRequestPayment(r, getRequestPaymentResponse)
		} else {
			ko.Spec.RequestPayment = nil
		}
	}

	// Tagging - directory buckets use S3 Control API
	if isDirectoryBucket {
		if err := rm.getDirectoryBucketTagging(ctx, r, ko); err != nil {
			return err
		}
	} else {
		getTaggingResponse, err := rm.sdkapi.GetBucketTagging(ctx, rm.newGetBucketTaggingPayload(r))
		if err != nil {
			if awsErr, ok := ackerr.AWSError(err); !ok || awsErr.ErrorCode() != "NoSuchTagSet" {
				return err
			}
		}
		if getTaggingResponse != nil && getTaggingResponse.TagSet != nil {
			ko.Spec.Tagging = rm.setResourceTagging(r, getTaggingResponse)
		} else {
			ko.Spec.Tagging = nil
		}
	}

	if !isDirectoryBucket {
		getVersioningResponse, err := rm.sdkapi.GetBucketVersioning(ctx, rm.newGetBucketVersioningPayload(r))
		if err != nil {
			return err
		}
		if getVersioningResponse.Status != "" {
			ko.Spec.Versioning = rm.setResourceVersioning(r, getVersioningResponse)
		} else {
			ko.Spec.Versioning = nil
		}

		getWebsiteResponse, err := rm.sdkapi.GetBucketWebsite(ctx, rm.newGetBucketWebsitePayload(r))
		if err != nil {
			if awsErr, ok := ackerr.AWSError(err); !ok || awsErr.ErrorCode() != "NoSuchWebsiteConfiguration" {
				return err
			}
		}
		if getWebsiteResponse != nil {
			ko.Spec.Website = rm.setResourceWebsite(r, getWebsiteResponse)
		} else {
			ko.Spec.Website = nil
		}
	}

	return nil
}

// customPreCompare ensures that default values of nil-able types are
// appropriately replaced with empty maps or structs depending on the default
// output of the SDK.
func customPreCompare(
	a *resource,
	b *resource,
) {
	if a.ko.Spec.Accelerate == nil && b.ko.Spec.Accelerate != nil {
		a.ko.Spec.Accelerate = &svcapitypes.AccelerateConfiguration{}

		if b.ko.Spec.Accelerate.Status != nil &&
			*b.ko.Spec.Accelerate.Status == string(DefaultAccelerationStatus) {
			a.ko.Spec.Accelerate.Status = aws.String(string(DefaultAccelerationStatus))
		}
	}
	if a.ko.Spec.Analytics == nil && b.ko.Spec.Analytics != nil {
		a.ko.Spec.Analytics = make([]*svcapitypes.AnalyticsConfiguration, 0)
	}
	if a.ko.Spec.ACL != nil {
		// Don't diff grant headers if a canned ACL has been used
		a.ko.Spec.GrantFullControl = nil
		b.ko.Spec.GrantFullControl = nil
		a.ko.Spec.GrantRead = nil
		b.ko.Spec.GrantRead = nil
		a.ko.Spec.GrantReadACP = nil
		b.ko.Spec.GrantReadACP = nil
		a.ko.Spec.GrantWrite = nil
		b.ko.Spec.GrantWrite = nil
		a.ko.Spec.GrantWriteACP = nil
		b.ko.Spec.GrantWriteACP = nil

		// Find the canned ACL from the joined possibility string
		if b.ko.Spec.ACL != nil {
			// Take the first ACL in case they are defined as delimited list
			a.ko.Spec.ACL = &strings.Split(*a.ko.Spec.ACL, CannedACLJoinDelimiter)[0]
			b.ko.Spec.ACL = matchPossibleCannedACL(*a.ko.Spec.ACL, *b.ko.Spec.ACL)
		}
	} else {
		// Ignore diff if possible canned ACLs are the default
		if b.ko.Spec.ACL != nil && isDefaultCannedACLPossibilities(*b.ko.Spec.ACL) {
			b.ko.Spec.ACL = nil
		}

		// If we are sure the grants weren't set from the header strings
		if a.ko.Spec.GrantFullControl == nil &&
			a.ko.Spec.GrantRead == nil &&
			a.ko.Spec.GrantReadACP == nil &&
			a.ko.Spec.GrantWrite == nil &&
			a.ko.Spec.GrantWriteACP == nil {
			b.ko.Spec.GrantFullControl = nil
			b.ko.Spec.GrantRead = nil
			b.ko.Spec.GrantReadACP = nil
			b.ko.Spec.GrantWrite = nil
			b.ko.Spec.GrantWriteACP = nil
		}

		emptyGrant := ""
		if a.ko.Spec.GrantFullControl == nil && b.ko.Spec.GrantFullControl != nil {
			a.ko.Spec.GrantFullControl = &emptyGrant
			// TODO(RedbackThomson): Remove the following line. GrantFullControl
			// has a server-side default of id="<owner ID>". This field needs to
			// be marked as such before we can diff it.
			b.ko.Spec.GrantFullControl = &emptyGrant
		}
		if a.ko.Spec.GrantRead == nil && b.ko.Spec.GrantRead != nil {
			a.ko.Spec.GrantRead = &emptyGrant
		}
		if a.ko.Spec.GrantReadACP == nil && b.ko.Spec.GrantReadACP != nil {
			a.ko.Spec.GrantReadACP = &emptyGrant
		}
		if a.ko.Spec.GrantWrite == nil && b.ko.Spec.GrantWrite != nil {
			a.ko.Spec.GrantWrite = &emptyGrant
		}
		if a.ko.Spec.GrantWriteACP == nil && b.ko.Spec.GrantWriteACP != nil {
			a.ko.Spec.GrantWriteACP = &emptyGrant
		}
	}

	if a.ko.Spec.CORS == nil && b.ko.Spec.CORS != nil {
		a.ko.Spec.CORS = &svcapitypes.CORSConfiguration{}
	}
	if a.ko.Spec.Encryption == nil && b.ko.Spec.Encryption != nil {
		a.ko.Spec.Encryption = &svcapitypes.ServerSideEncryptionConfiguration{}
	}
	if a.ko.Spec.IntelligentTiering == nil && b.ko.Spec.IntelligentTiering != nil {
		a.ko.Spec.IntelligentTiering = make([]*svcapitypes.IntelligentTieringConfiguration, 0)
	}
	if a.ko.Spec.Inventory == nil && b.ko.Spec.Inventory != nil {
		a.ko.Spec.Inventory = make([]*svcapitypes.InventoryConfiguration, 0)
	}
	if a.ko.Spec.Lifecycle == nil && b.ko.Spec.Lifecycle != nil {
		a.ko.Spec.Lifecycle = &svcapitypes.BucketLifecycleConfiguration{}
	}
	if a.ko.Spec.Logging == nil && b.ko.Spec.Logging != nil {
		a.ko.Spec.Logging = &svcapitypes.BucketLoggingStatus{}
	}
	if a.ko.Spec.Metrics == nil && b.ko.Spec.Metrics != nil {
		a.ko.Spec.Metrics = make([]*svcapitypes.MetricsConfiguration, 0)
	}
	if a.ko.Spec.Notification == nil && b.ko.Spec.Notification != nil {
		a.ko.Spec.Notification = &svcapitypes.NotificationConfiguration{}
	}
	if a.ko.Spec.OwnershipControls == nil && b.ko.Spec.OwnershipControls != nil {
		a.ko.Spec.OwnershipControls = &svcapitypes.OwnershipControls{}
	}
	if a.ko.Spec.PublicAccessBlock == nil && b.ko.Spec.PublicAccessBlock != nil {
		a.ko.Spec.PublicAccessBlock = &DefaultPublicBlockAccess
	}
	if a.ko.Spec.Replication == nil && b.ko.Spec.Replication != nil {
		a.ko.Spec.Replication = &svcapitypes.ReplicationConfiguration{}
	}
	if a.ko.Spec.RequestPayment == nil && b.ko.Spec.RequestPayment != nil {
		a.ko.Spec.RequestPayment = &svcapitypes.RequestPaymentConfiguration{
			Payer: aws.String(string(DefaultRequestPayer)),
		}
	}
	if a.ko.Spec.Tagging == nil && b.ko.Spec.Tagging != nil {
		a.ko.Spec.Tagging = &svcapitypes.Tagging{}
	}
	if a.ko.Spec.Versioning == nil && b.ko.Spec.Versioning != nil {
		a.ko.Spec.Versioning = &svcapitypes.VersioningConfiguration{}

		if b.ko.Spec.Versioning.Status != nil &&
			*b.ko.Spec.Versioning.Status == string(DefaultVersioningStatus) {
			a.ko.Spec.Versioning.Status = aws.String(string(DefaultVersioningStatus))
		}
	}
	if a.ko.Spec.Website == nil && b.ko.Spec.Website != nil {
		a.ko.Spec.Website = &svcapitypes.WebsiteConfiguration{}
	}
}

//region accelerate

func (rm *resourceManager) newGetBucketAcceleratePayload(
	r *resource,
) *svcsdk.GetBucketAccelerateConfigurationInput {
	res := &svcsdk.GetBucketAccelerateConfigurationInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) newPutBucketAcceleratePayload(
	r *resource,
) *svcsdk.PutBucketAccelerateConfigurationInput {
	res := &svcsdk.PutBucketAccelerateConfigurationInput{}
	res.Bucket = r.ko.Spec.Name
	if r.ko.Spec.Accelerate != nil {
		res.AccelerateConfiguration = rm.newAccelerateConfiguration(r)
	} else {
		res.AccelerateConfiguration = &svcsdktypes.AccelerateConfiguration{}
	}

	if res.AccelerateConfiguration.Status == "" {
		res.AccelerateConfiguration.Status = DefaultAccelerationStatus
	}

	return res
}

func (rm *resourceManager) syncAccelerate(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncAccelerate")
	defer exit(err)
	input := rm.newPutBucketAcceleratePayload(r)

	_, err = rm.sdkapi.PutBucketAccelerateConfiguration(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketAccelerate", err)
	if err != nil {
		return err
	}

	return nil
}

//endregion accelerate

//region acl

// setResourceACL sets the `Grant*` spec fields given the output of a
// `GetBucketAcl` operation.
func (rm *resourceManager) setResourceACL(
	ko *svcapitypes.Bucket,
	resp *svcsdk.GetBucketAclOutput,
) {
	grants := GetHeadersFromGrants(resp)
	if grants.FullControl != "" {
		ko.Spec.GrantFullControl = &grants.FullControl
	}
	if grants.Read != "" {
		ko.Spec.GrantRead = &grants.Read
	}
	if grants.ReadACP != "" {
		ko.Spec.GrantReadACP = &grants.ReadACP
	}
	if grants.Write != "" {
		ko.Spec.GrantWrite = &grants.Write
	}
	if grants.WriteACP != "" {
		ko.Spec.GrantWriteACP = &grants.WriteACP
	}

	// Join possible ACLs into a single string, delimited by bar
	cannedACLs := GetPossibleCannedACLsFromGrants(resp)
	joinedACLs := strings.Join(cannedACLs, CannedACLJoinDelimiter)
	if joinedACLs != "" {
		ko.Spec.ACL = &joinedACLs
	}
}

func (rm *resourceManager) newGetBucketACLPayload(
	r *resource,
) *svcsdk.GetBucketAclInput {
	res := &svcsdk.GetBucketAclInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) newPutBucketACLPayload(
	r *resource,
) *svcsdk.PutBucketAclInput {
	res := &svcsdk.PutBucketAclInput{}
	res.Bucket = r.ko.Spec.Name
	if r.ko.Spec.ACL != nil {
		res.ACL = svcsdktypes.BucketCannedACL(*r.ko.Spec.ACL)
	} else {
		// Only put grants on bucket if there is no canned ACL to match

		if r.ko.Spec.GrantFullControl != nil {
			res.GrantFullControl = r.ko.Spec.GrantFullControl
		}
		if r.ko.Spec.GrantRead != nil {
			res.GrantRead = r.ko.Spec.GrantRead
		}
		if r.ko.Spec.GrantReadACP != nil {
			res.GrantReadACP = r.ko.Spec.GrantReadACP
		}
		if r.ko.Spec.GrantWrite != nil {
			res.GrantWrite = r.ko.Spec.GrantWrite
		}
		if r.ko.Spec.GrantWriteACP != nil {
			res.GrantWriteACP = r.ko.Spec.GrantWriteACP
		}
	}

	// Check that there is at least some ACL on the bucket
	if res.ACL == "" &&
		res.GrantFullControl == nil &&
		res.GrantRead == nil &&
		res.GrantReadACP == nil &&
		res.GrantWrite == nil &&
		res.GrantWriteACP == nil {
		res.ACL = DefaultACL
	}

	return res
}

func (rm *resourceManager) syncACL(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncACL")
	defer exit(err)
	input := rm.newPutBucketACLPayload(r)

	_, err = rm.sdkapi.PutBucketAcl(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketAcl", err)
	if err != nil {
		return err
	}

	return nil
}

//endregion acl

//region cors

func (rm *resourceManager) newGetBucketCORSPayload(
	r *resource,
) *svcsdk.GetBucketCorsInput {
	res := &svcsdk.GetBucketCorsInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) newPutBucketCORSPayload(
	r *resource,
) *svcsdk.PutBucketCorsInput {
	res := &svcsdk.PutBucketCorsInput{}
	res.Bucket = r.ko.Spec.Name
	res.CORSConfiguration = rm.newCORSConfiguration(r)

	if res.CORSConfiguration.CORSRules == nil {
		res.CORSConfiguration.CORSRules = []svcsdktypes.CORSRule{}
	}

	return res
}

func (rm *resourceManager) newDeleteBucketCORSPayload(
	r *resource,
) *svcsdk.DeleteBucketCorsInput {
	res := &svcsdk.DeleteBucketCorsInput{}
	res.Bucket = r.ko.Spec.Name

	return res
}

func (rm *resourceManager) putCORS(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.putCORS")
	defer exit(err)
	input := rm.newPutBucketCORSPayload(r)

	_, err = rm.sdkapi.PutBucketCors(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketCors", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) deleteCORS(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.deleteCORS")
	defer exit(err)
	input := rm.newDeleteBucketCORSPayload(r)

	_, err = rm.sdkapi.DeleteBucketCors(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeleteBucketCors", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) syncCORS(
	ctx context.Context,
	r *resource,
) (err error) {
	if r.ko.Spec.CORS == nil || r.ko.Spec.CORS.CORSRules == nil {
		return rm.deleteCORS(ctx, r)
	}
	return rm.putCORS(ctx, r)
}

//endregion cors

//region encryption

func (rm *resourceManager) newGetBucketEncryptionPayload(
	r *resource,
) *svcsdk.GetBucketEncryptionInput {
	res := &svcsdk.GetBucketEncryptionInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) newPutBucketEncryptionPayload(
	r *resource,
) *svcsdk.PutBucketEncryptionInput {
	res := &svcsdk.PutBucketEncryptionInput{}
	res.Bucket = r.ko.Spec.Name
	res.ServerSideEncryptionConfiguration = rm.newServerSideEncryptionConfiguration(r)

	return res
}

func (rm *resourceManager) newDeleteBucketEncryptionPayload(
	r *resource,
) *svcsdk.DeleteBucketEncryptionInput {
	res := &svcsdk.DeleteBucketEncryptionInput{}
	res.Bucket = r.ko.Spec.Name

	return res
}

func (rm *resourceManager) putEncryption(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.putEncryption")
	defer exit(err)
	input := rm.newPutBucketEncryptionPayload(r)

	_, err = rm.sdkapi.PutBucketEncryption(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketEncryption", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) deleteEncryption(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.deleteEncryption")
	defer exit(err)
	input := rm.newDeleteBucketEncryptionPayload(r)

	_, err = rm.sdkapi.DeleteBucketEncryption(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeleteBucketEncryption", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) syncEncryption(
	ctx context.Context,
	r *resource,
) (err error) {
	if r.ko.Spec.Encryption == nil || r.ko.Spec.Encryption.Rules == nil {
		return rm.deleteEncryption(ctx, r)
	}
	return rm.putEncryption(ctx, r)
}

//endregion encryption

//region lifecycle

func (rm *resourceManager) newGetBucketLifecyclePayload(
	r *resource,
) *svcsdk.GetBucketLifecycleConfigurationInput {
	res := &svcsdk.GetBucketLifecycleConfigurationInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) newPutBucketLifecyclePayload(
	r *resource,
) *svcsdk.PutBucketLifecycleConfigurationInput {
	res := &svcsdk.PutBucketLifecycleConfigurationInput{}
	res.Bucket = r.ko.Spec.Name
	res.LifecycleConfiguration = rm.newLifecycleConfiguration(r)
	return res
}

func (rm *resourceManager) newDeleteBucketLifecyclePayload(
	r *resource,
) *svcsdk.DeleteBucketLifecycleInput {
	res := &svcsdk.DeleteBucketLifecycleInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) putLifecycle(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.putLifecycle")
	defer exit(err)
	input := rm.newPutBucketLifecyclePayload(r)

	_, err = rm.sdkapi.PutBucketLifecycleConfiguration(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketLifecycle", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) deleteLifecycle(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.deleteLifecycle")
	defer exit(err)
	input := rm.newDeleteBucketLifecyclePayload(r)

	_, err = rm.sdkapi.DeleteBucketLifecycle(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeleteBucketLifecycle", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) syncLifecycle(
	ctx context.Context,
	r *resource,
) (err error) {
	if r.ko.Spec.Lifecycle == nil || r.ko.Spec.Lifecycle.Rules == nil {
		return rm.deleteLifecycle(ctx, r)
	}
	return rm.putLifecycle(ctx, r)
}

//endregion lifecycle

//region logging

func (rm *resourceManager) newGetBucketLoggingPayload(
	r *resource,
) *svcsdk.GetBucketLoggingInput {
	res := &svcsdk.GetBucketLoggingInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) newPutBucketLoggingPayload(
	r *resource,
) *svcsdk.PutBucketLoggingInput {
	res := &svcsdk.PutBucketLoggingInput{}
	res.Bucket = r.ko.Spec.Name
	if r.ko.Spec.Logging != nil {
		res.BucketLoggingStatus = rm.newBucketLoggingStatus(r)
	} else {
		res.BucketLoggingStatus = &svcsdktypes.BucketLoggingStatus{}
	}
	return res
}

func (rm *resourceManager) syncLogging(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncLogging")
	defer exit(err)
	input := rm.newPutBucketLoggingPayload(r)

	_, err = rm.sdkapi.PutBucketLogging(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketLogging", err)
	if err != nil {
		return err
	}

	return nil
}

//endregion logging

//region notification

func (rm *resourceManager) newGetBucketNotificationPayload(
	r *resource,
) *svcsdk.GetBucketNotificationConfigurationInput {
	res := &svcsdk.GetBucketNotificationConfigurationInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) newPutBucketNotificationPayload(
	r *resource,
) *svcsdk.PutBucketNotificationConfigurationInput {
	res := &svcsdk.PutBucketNotificationConfigurationInput{}
	res.Bucket = r.ko.Spec.Name
	if r.ko.Spec.Notification != nil {
		res.NotificationConfiguration = rm.newNotificationConfiguration(r)
	} else {
		res.NotificationConfiguration = &svcsdktypes.NotificationConfiguration{}
	}
	return res
}

func (rm *resourceManager) syncNotification(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncNotification")
	defer exit(err)
	input := rm.newPutBucketNotificationPayload(r)

	_, err = rm.sdkapi.PutBucketNotificationConfiguration(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketNotification", err)
	if err != nil {
		return err
	}

	return nil
}

//endregion notification

//region ownershipcontrols

func (rm *resourceManager) newGetBucketOwnershipControlsPayload(
	r *resource,
) *svcsdk.GetBucketOwnershipControlsInput {
	res := &svcsdk.GetBucketOwnershipControlsInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) newPutBucketOwnershipControlsPayload(
	r *resource,
) *svcsdk.PutBucketOwnershipControlsInput {
	res := &svcsdk.PutBucketOwnershipControlsInput{}
	res.Bucket = r.ko.Spec.Name
	res.OwnershipControls = rm.newOwnershipControls(r)

	return res
}

func (rm *resourceManager) newDeleteBucketOwnershipControlsPayload(
	r *resource,
) *svcsdk.DeleteBucketOwnershipControlsInput {
	res := &svcsdk.DeleteBucketOwnershipControlsInput{}
	res.Bucket = r.ko.Spec.Name

	return res
}

func (rm *resourceManager) putOwnershipControls(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.putOwnershipControls")
	defer exit(err)
	input := rm.newPutBucketOwnershipControlsPayload(r)

	_, err = rm.sdkapi.PutBucketOwnershipControls(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketOwnershipControls", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) deleteOwnershipControls(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.deleteOwnershipControls")
	defer exit(err)
	input := rm.newDeleteBucketOwnershipControlsPayload(r)

	_, err = rm.sdkapi.DeleteBucketOwnershipControls(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeleteBucketOwnershipControls", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) syncOwnershipControls(
	ctx context.Context,
	r *resource,
) (err error) {
	if r.ko.Spec.OwnershipControls == nil || r.ko.Spec.OwnershipControls.Rules == nil {
		return rm.deleteOwnershipControls(ctx, r)
	}
	return rm.putOwnershipControls(ctx, r)
}

//endregion ownershipcontrols

//region policy

func (rm *resourceManager) newGetBucketPolicyPayload(
	r *resource,
) *svcsdk.GetBucketPolicyInput {
	res := &svcsdk.GetBucketPolicyInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) newPutBucketPolicyPayload(
	r *resource,
	isDirectoryBucket bool,
) *svcsdk.PutBucketPolicyInput {
	res := &svcsdk.PutBucketPolicyInput{}
	res.Bucket = r.ko.Spec.Name
	res.Policy = r.ko.Spec.Policy

	// ConfirmRemoveSelfBucketAccess is not supported by directory
	if !isDirectoryBucket {
		res.ConfirmRemoveSelfBucketAccess = aws.Bool(false)
	}

	return res
}

func (rm *resourceManager) newDeleteBucketPolicyPayload(
	r *resource,
) *svcsdk.DeleteBucketPolicyInput {
	res := &svcsdk.DeleteBucketPolicyInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) putPolicy(
	ctx context.Context,
	r *resource,
	isDirectoryBucket bool,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.putPolicy")
	defer exit(err)
	input := rm.newPutBucketPolicyPayload(r, isDirectoryBucket)

	_, err = rm.sdkapi.PutBucketPolicy(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketPolicy", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) deletePolicy(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.deletePolicy")
	defer exit(err)
	input := rm.newDeleteBucketPolicyPayload(r)

	_, err = rm.sdkapi.DeleteBucketPolicy(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeleteBucketPolicy", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) syncPolicy(
	ctx context.Context,
	r *resource,
	isDirectoryBucket bool,
) (err error) {
	if r.ko.Spec.Policy == nil || *r.ko.Spec.Policy == "" {
		return rm.deletePolicy(ctx, r)
	}
	return rm.putPolicy(ctx, r, isDirectoryBucket)
}

//endregion

//region publicaccessblock

func (rm *resourceManager) newGetPublicAccessBlockPayload(
	r *resource,
) *svcsdk.GetPublicAccessBlockInput {
	res := &svcsdk.GetPublicAccessBlockInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) newPutPublicAccessBlockPayload(
	r *resource,
) *svcsdk.PutPublicAccessBlockInput {
	res := &svcsdk.PutPublicAccessBlockInput{}
	res.Bucket = r.ko.Spec.Name
	res.PublicAccessBlockConfiguration = rm.newPublicAccessBlockConfiguration(r)

	return res
}

func (rm *resourceManager) newDeletePublicAccessBlockPayload(
	r *resource,
) *svcsdk.DeletePublicAccessBlockInput {
	res := &svcsdk.DeletePublicAccessBlockInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) putPublicAccessBlock(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.putPublicAccessBlock")
	defer exit(err)
	input := rm.newPutPublicAccessBlockPayload(r)

	_, err = rm.sdkapi.PutPublicAccessBlock(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutPublicAccessBlock", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) deletePublicAccessBlock(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.deletePublicAccessBlock")
	defer exit(err)
	input := rm.newDeletePublicAccessBlockPayload(r)

	_, err = rm.sdkapi.DeletePublicAccessBlock(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeletePublicAccessBlock", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) syncPublicAccessBlock(
	ctx context.Context,
	r *resource,
) (err error) {
	if r.ko.Spec.PublicAccessBlock == nil || *r.ko.Spec.PublicAccessBlock == DefaultPublicBlockAccess {
		return rm.deletePublicAccessBlock(ctx, r)
	}
	return rm.putPublicAccessBlock(ctx, r)
}

//endregion

//region replication

func (rm *resourceManager) newGetBucketReplicationPayload(
	r *resource,
) *svcsdk.GetBucketReplicationInput {
	res := &svcsdk.GetBucketReplicationInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) newPutBucketReplicationPayload(
	r *resource,
) *svcsdk.PutBucketReplicationInput {
	res := &svcsdk.PutBucketReplicationInput{}
	res.Bucket = r.ko.Spec.Name
	res.ReplicationConfiguration = rm.newReplicationConfiguration(r)
	return res
}

func (rm *resourceManager) newDeleteBucketReplicationPayload(
	r *resource,
) *svcsdk.DeleteBucketReplicationInput {
	res := &svcsdk.DeleteBucketReplicationInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) putReplication(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.putReplication")
	defer exit(err)
	input := rm.newPutBucketReplicationPayload(r)

	_, err = rm.sdkapi.PutBucketReplication(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketReplication", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) deleteReplication(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.deleteReplication")
	defer exit(err)
	input := rm.newDeleteBucketReplicationPayload(r)

	_, err = rm.sdkapi.DeleteBucketReplication(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeleteBucketReplication", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) syncReplication(
	ctx context.Context,
	r *resource,
) (err error) {
	if r.ko.Spec.Replication == nil || r.ko.Spec.Replication.Rules == nil {
		return rm.deleteReplication(ctx, r)
	}
	return rm.putReplication(ctx, r)
}

//endregion replication

//region requestpayment

func (rm *resourceManager) newGetBucketRequestPaymentPayload(
	r *resource,
) *svcsdk.GetBucketRequestPaymentInput {
	res := &svcsdk.GetBucketRequestPaymentInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) newPutBucketRequestPaymentPayload(
	r *resource,
) *svcsdk.PutBucketRequestPaymentInput {
	res := &svcsdk.PutBucketRequestPaymentInput{}
	res.Bucket = r.ko.Spec.Name
	if r.ko.Spec.RequestPayment != nil && r.ko.Spec.RequestPayment.Payer != nil {
		res.RequestPaymentConfiguration = rm.newRequestPaymentConfiguration(r)
	} else {
		res.RequestPaymentConfiguration = &svcsdktypes.RequestPaymentConfiguration{}
	}

	if res.RequestPaymentConfiguration.Payer == "" {
		res.RequestPaymentConfiguration.Payer = DefaultRequestPayer
	}

	return res
}

func (rm *resourceManager) syncRequestPayment(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncRequestPayment")
	defer exit(err)
	input := rm.newPutBucketRequestPaymentPayload(r)

	_, err = rm.sdkapi.PutBucketRequestPayment(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketRequestPayment", err)
	if err != nil {
		return err
	}

	return nil
}

//endregion requestpayment

//region tagging

func (rm *resourceManager) newGetBucketTaggingPayload(
	r *resource,
) *svcsdk.GetBucketTaggingInput {
	res := &svcsdk.GetBucketTaggingInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) newPutBucketTaggingPayload(
	r *resource,
) *svcsdk.PutBucketTaggingInput {
	res := &svcsdk.PutBucketTaggingInput{}
	res.Bucket = r.ko.Spec.Name
	res.Tagging = rm.newTagging(r)

	return res
}

func (rm *resourceManager) newDeleteBucketTaggingPayload(
	r *resource,
) *svcsdk.DeleteBucketTaggingInput {
	res := &svcsdk.DeleteBucketTaggingInput{}
	res.Bucket = r.ko.Spec.Name

	return res
}

func (rm *resourceManager) putTagging(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.putTagging")
	defer exit(err)

	// Directory buckets use S3 Control API for tagging
	if r.ko.Spec.Name != nil && IsDirectoryBucketName(*r.ko.Spec.Name) {
		return rm.putDirectoryBucketTagging(ctx, r)
	}

	input := rm.newPutBucketTaggingPayload(r)

	_, err = rm.sdkapi.PutBucketTagging(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketTagging", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) deleteTagging(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.deleteTagging")
	defer exit(err)

	// Directory buckets use S3 Control API for tagging
	if r.ko.Spec.Name != nil && IsDirectoryBucketName(*r.ko.Spec.Name) {
		return rm.deleteDirectoryBucketTagging(ctx, r)
	}

	input := rm.newDeleteBucketTaggingPayload(r)

	_, err = rm.sdkapi.DeleteBucketTagging(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeleteBucketTagging", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) syncTagging(
	ctx context.Context,
	r *resource,
) (err error) {
	if r.ko.Spec.Tagging == nil || r.ko.Spec.Tagging.TagSet == nil {
		return rm.deleteTagging(ctx, r)
	}
	return rm.putTagging(ctx, r)
}

// For Tagging operations, directory buckets require the S3 Control API
// (TagResource, ListTagsForResource, UntagResource) instead of the standard S3 tagging APIs
// (PutBucketTagging, GetBucketTagging, DeleteBucketTagging).
// This is because directory buckets are addressed by ARN.

func (rm *resourceManager) putDirectoryBucketTagging(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.putDirectoryBucketTagging")
	defer exit(err)

	s3controlClient := s3control.NewFromConfig(rm.clientcfg)

	desiredKeys := make(map[string]struct{})
	var desiredTags []s3controltypes.Tag
	if r.ko.Spec.Tagging != nil && r.ko.Spec.Tagging.TagSet != nil {
		for _, tag := range r.ko.Spec.Tagging.TagSet {
			if tag == nil || tag.Key == nil || tag.Value == nil {
				continue
			}
			desiredKeys[*tag.Key] = struct{}{}
			desiredTags = append(desiredTags, s3controltypes.Tag{
				Key:   tag.Key,
				Value: tag.Value,
			})
		}
	}

	// Status ARN may not be set yet during create
	var bucketARN string
	if r.ko.Status.ACKResourceMetadata != nil && r.ko.Status.ACKResourceMetadata.ARN != nil {
		bucketARN = string(*r.ko.Status.ACKResourceMetadata.ARN)
	} else {
		bucketARN = rm.directoryBucketARN(*r.ko.Spec.Name)
	}
	accountID := aws.String(string(rm.awsAccountID))

	// S3Control TagResource only adds/updates tags. To reconcile fully,
	// first remove tags absent from the desired set.
	listInput := &s3control.ListTagsForResourceInput{
		ResourceArn: aws.String(bucketARN),
		AccountId:   accountID,
	}
	listResp, err := s3controlClient.ListTagsForResource(ctx, listInput)
	rm.metrics.RecordAPICall("READ", "ListTagsForResource", err)
	if err != nil {
		if awsErr, ok := ackerr.AWSError(err); !ok || awsErr.ErrorCode() != "NoSuchTagSet" {
			return err
		}
	}

	var keysToRemove []string
	if listResp != nil {
		for _, tag := range listResp.Tags {
			if tag.Key == nil {
				continue
			}
			if _, ok := desiredKeys[*tag.Key]; !ok {
				keysToRemove = append(keysToRemove, *tag.Key)
			}
		}
	}
	if len(keysToRemove) > 0 {
		untagInput := &s3control.UntagResourceInput{
			ResourceArn: aws.String(bucketARN),
			AccountId:   accountID,
			TagKeys:     keysToRemove,
		}
		_, err = s3controlClient.UntagResource(ctx, untagInput)
		rm.metrics.RecordAPICall("DELETE", "UntagResource", err)
		if err != nil {
			return err
		}
	}

	if len(desiredTags) > 0 {
		tagInput := &s3control.TagResourceInput{
			ResourceArn: aws.String(bucketARN),
			AccountId:   accountID,
			Tags:        desiredTags,
		}
		_, err = s3controlClient.TagResource(ctx, tagInput)
		rm.metrics.RecordAPICall("UPDATE", "TagResource", err)
		if err != nil {
			return err
		}
	}

	return nil
}

// getDirectoryBucketTagging uses S3 Control API to get tags for directory buckets
func (rm *resourceManager) getDirectoryBucketTagging(
	ctx context.Context,
	r *resource,
	ko *svcapitypes.Bucket,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.getDirectoryBucketTagging")
	defer func() { exit(err) }()

	s3controlClient := s3control.NewFromConfig(rm.clientcfg)

	// Status ARN may not be set yet during create
	var bucketARN string
	if r.ko.Status.ACKResourceMetadata != nil && r.ko.Status.ACKResourceMetadata.ARN != nil {
		bucketARN = string(*r.ko.Status.ACKResourceMetadata.ARN)
	} else {
		bucketARN = rm.directoryBucketARN(*r.ko.Spec.Name)
	}
	input := &s3control.ListTagsForResourceInput{
		ResourceArn: aws.String(bucketARN),
		AccountId:   aws.String(string(rm.awsAccountID)),
	}

	resp, err := s3controlClient.ListTagsForResource(ctx, input)
	rm.metrics.RecordAPICall("READ", "ListTagsForResource", err)
	if err != nil {
		if awsErr, ok := ackerr.AWSError(err); ok && awsErr.ErrorCode() == "NoSuchTagSet" {
			ko.Spec.Tagging = nil
			return nil
		}
		return err
	}

	if resp != nil && len(resp.Tags) > 0 {
		ko.Spec.Tagging = &svcapitypes.Tagging{
			TagSet: make([]*svcapitypes.Tag, len(resp.Tags)),
		}
		for i, tag := range resp.Tags {
			ko.Spec.Tagging.TagSet[i] = &svcapitypes.Tag{
				Key:   tag.Key,
				Value: tag.Value,
			}
		}
	} else {
		ko.Spec.Tagging = nil
	}

	return nil
}

// deleteDirectoryBucketTagging uses S3 Control API to delete tags for directory buckets
func (rm *resourceManager) deleteDirectoryBucketTagging(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.deleteDirectoryBucketTagging")
	defer exit(err)

	s3controlClient := s3control.NewFromConfig(rm.clientcfg)

	bucketARN := string(*r.ko.Status.ACKResourceMetadata.ARN)
	getInput := &s3control.ListTagsForResourceInput{
		ResourceArn: aws.String(bucketARN),
		AccountId:   aws.String(string(rm.awsAccountID)),
	}

	resp, err := s3controlClient.ListTagsForResource(ctx, getInput)
	rm.metrics.RecordAPICall("READ", "ListTagsForResource", err)
	if err != nil {
		if awsErr, ok := ackerr.AWSError(err); ok && awsErr.ErrorCode() == "NoSuchTagSet" {
			return nil
		}
		return err
	}

	if len(resp.Tags) == 0 {
		return nil
	}

	var tagKeys []string
	for _, tag := range resp.Tags {
		if tag.Key != nil {
			tagKeys = append(tagKeys, *tag.Key)
		}
	}

	deleteInput := &s3control.UntagResourceInput{
		ResourceArn: aws.String(bucketARN),
		AccountId:   aws.String(string(rm.awsAccountID)),
		TagKeys:     tagKeys,
	}

	_, err = s3controlClient.UntagResource(ctx, deleteInput)
	rm.metrics.RecordAPICall("DELETE", "UntagResource", err)
	if err != nil {
		return err
	}

	return nil
}

//endregion tagging

//region versioning

func (rm *resourceManager) newGetBucketVersioningPayload(
	r *resource,
) *svcsdk.GetBucketVersioningInput {
	res := &svcsdk.GetBucketVersioningInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) newPutBucketVersioningPayload(
	r *resource,
) *svcsdk.PutBucketVersioningInput {
	res := &svcsdk.PutBucketVersioningInput{}
	res.Bucket = r.ko.Spec.Name
	if r.ko.Spec.Versioning != nil {
		res.VersioningConfiguration = rm.newVersioningConfiguration(r)
	} else {
		res.VersioningConfiguration = &svcsdktypes.VersioningConfiguration{}
	}

	if res.VersioningConfiguration.Status == "" {
		res.VersioningConfiguration.Status = DefaultVersioningStatus
	}

	return res
}

func (rm *resourceManager) syncVersioning(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncVersioning")
	defer exit(err)
	input := rm.newPutBucketVersioningPayload(r)

	_, err = rm.sdkapi.PutBucketVersioning(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketVersioning", err)
	if err != nil {
		return err
	}

	return nil
}

//endregion versioning

//region website

func (rm *resourceManager) newGetBucketWebsitePayload(
	r *resource,
) *svcsdk.GetBucketWebsiteInput {
	res := &svcsdk.GetBucketWebsiteInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) newPutBucketWebsitePayload(
	r *resource,
) *svcsdk.PutBucketWebsiteInput {
	res := &svcsdk.PutBucketWebsiteInput{}
	res.Bucket = r.ko.Spec.Name
	res.WebsiteConfiguration = rm.newWebsiteConfiguration(r)

	return res
}

func (rm *resourceManager) newDeleteBucketWebsitePayload(
	r *resource,
) *svcsdk.DeleteBucketWebsiteInput {
	res := &svcsdk.DeleteBucketWebsiteInput{}
	res.Bucket = r.ko.Spec.Name

	return res
}

func (rm *resourceManager) putWebsite(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.putWebsite")
	defer exit(err)
	input := rm.newPutBucketWebsitePayload(r)

	_, err = rm.sdkapi.PutBucketWebsite(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketWebsite", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) deleteWebsite(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.deleteWebsite")
	defer exit(err)
	input := rm.newDeleteBucketWebsitePayload(r)

	_, err = rm.sdkapi.DeleteBucketWebsite(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeleteBucketWebsite", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) syncWebsite(
	ctx context.Context,
	r *resource,
) (err error) {
	// IndexDocument is a required field for PutBucketWebsite API call. If it is missing
	// we want to call DeleteBucketWebsite
	// https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketWebsite.html
	if r.ko.Spec.Website == nil || r.ko.Spec.Website.IndexDocument == nil {
		return rm.deleteWebsite(ctx, r)
	}
	return rm.putWebsite(ctx, r)
}

//endregion website
