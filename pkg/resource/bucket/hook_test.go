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
	"errors"
	"testing"

	smithy "github.com/aws/smithy-go"
	smithymiddleware "github.com/aws/smithy-go/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	svcsdk "github.com/aws/aws-sdk-go-v2/service/s3"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/s3/types"

	svcapitypes "github.com/aws-controllers-k8s/s3-controller/apis/v1alpha1"
)

// opResult is a canned response for a single S3 operation. Exactly one of
// output or err is used; when err is non-nil it is returned to the caller as
// the operation's error.
type opResult struct {
	output interface{}
	err    error
}

// apiErr returns a smithy API error with the given code, matching how the
// runtime's ackerr.AWSError unwraps SDK errors.
func apiErr(code string) error {
	return &smithy.GenericAPIError{Code: code}
}

// newMockedSDKClient builds a real *s3.Client whose middleware stack is
// short-circuited at the Finalize step. By the time Finalize runs the smithy
// runtime has already recorded the operation name on the context, so we can
// dispatch a canned response per operation without performing any endpoint
// resolution, request signing or network I/O.
//
// The defaults model a freshly-created bucket with no optional configuration:
// the "Get*" operations whose property is unset return the not-found-style API
// error that addPutFieldsToSpec is written to ignore, and the rest return an
// empty output. Tests override individual operations via results.
func newMockedSDKClient(results map[string]opResult) *svcsdk.Client {
	defaults := map[string]opResult{
		"GetBucketAccelerateConfiguration":           {output: &svcsdk.GetBucketAccelerateConfigurationOutput{}},
		"ListBucketAnalyticsConfigurations":          {output: &svcsdk.ListBucketAnalyticsConfigurationsOutput{}},
		"GetBucketAcl":                               {output: &svcsdk.GetBucketAclOutput{Owner: &svcsdktypes.Owner{}}},
		"GetBucketCors":                              {err: apiErr("NoSuchCORSConfiguration")},
		"GetBucketEncryption":                        {output: &svcsdk.GetBucketEncryptionOutput{ServerSideEncryptionConfiguration: &svcsdktypes.ServerSideEncryptionConfiguration{}}},
		"ListBucketIntelligentTieringConfigurations": {output: &svcsdk.ListBucketIntelligentTieringConfigurationsOutput{}},
		"ListBucketInventoryConfigurations":          {output: &svcsdk.ListBucketInventoryConfigurationsOutput{}},
		"GetBucketLifecycleConfiguration":            {err: apiErr("NoSuchLifecycleConfiguration")},
		"GetBucketLogging":                           {output: &svcsdk.GetBucketLoggingOutput{}},
		"ListBucketMetricsConfigurations":            {output: &svcsdk.ListBucketMetricsConfigurationsOutput{}},
		"GetBucketNotificationConfiguration":         {output: &svcsdk.GetBucketNotificationConfigurationOutput{}},
		"GetBucketOwnershipControls":                 {err: apiErr("OwnershipControlsNotFoundError")},
		"GetBucketPolicy":                            {err: apiErr("NoSuchBucketPolicy")},
		"GetPublicAccessBlock":                       {err: apiErr("NoSuchPublicAccessBlockConfiguration")},
		"GetBucketReplication":                       {err: apiErr("ReplicationConfigurationNotFoundError")},
		"GetBucketRequestPayment":                    {output: &svcsdk.GetBucketRequestPaymentOutput{}},
		"GetBucketTagging":                           {err: apiErr("NoSuchTagSet")},
		"GetBucketVersioning":                        {output: &svcsdk.GetBucketVersioningOutput{}},
		"GetBucketWebsite":                           {err: apiErr("NoSuchWebsiteConfiguration")},
		"GetObjectLockConfiguration":                 {err: apiErr("ObjectLockConfigurationNotFoundError")},
	}

	mockFinalize := smithymiddleware.FinalizeMiddlewareFunc(
		"mockS3Finalize",
		func(
			ctx context.Context,
			in smithymiddleware.FinalizeInput,
			_ smithymiddleware.FinalizeHandler,
		) (smithymiddleware.FinalizeOutput, smithymiddleware.Metadata, error) {
			opName := smithymiddleware.GetOperationName(ctx)
			res, ok := results[opName]
			if !ok {
				res = defaults[opName]
			}
			return smithymiddleware.FinalizeOutput{Result: res.output}, smithymiddleware.Metadata{}, res.err
		},
	)

	return svcsdk.New(svcsdk.Options{
		Region: "us-west-2",
		APIOptions: []func(*smithymiddleware.Stack) error{
			func(stack *smithymiddleware.Stack) error {
				return stack.Finalize.Add(mockFinalize, smithymiddleware.Before)
			},
		},
	})
}

func newBucketResource(name string) *resource {
	return &resource{
		ko: &svcapitypes.Bucket{
			Spec: svcapitypes.BucketSpec{
				Name: &name,
			},
		},
	}
}

// Test_addPutFieldsToSpec_RequestPaymentError is a regression test for
// aws-controllers-k8s/community#2935 and #2936. The GetBucketRequestPayment
// error handler used to `return nil`, silently exiting addPutFieldsToSpec
// before it read Tagging, Versioning and Website back from AWS. As a result
// the observed (latest) resource kept the user's desired values for those
// fields, the delta found no difference and the controller never issued the
// Put* calls to apply them. The fix returns the error so the read fails loudly
// and the reconcile is retried.
func Test_addPutFieldsToSpec_RequestPaymentError(t *testing.T) {
	require := require.New(t)

	requestPaymentErr := errors.New("AccessDenied: not authorized to perform s3:GetBucketRequestPayment")

	rm := &resourceManager{
		sdkapi: newMockedSDKClient(map[string]opResult{
			"GetBucketRequestPayment": {err: requestPaymentErr},
			// These succeed in AWS but must never be reached on the error path,
			// and must never be reached *before* the error surfaces either.
			"GetBucketTagging": {output: &svcsdk.GetBucketTaggingOutput{
				TagSet: []svcsdktypes.Tag{{Key: strPtr("k"), Value: strPtr("v")}},
			}},
			"GetBucketVersioning": {output: &svcsdk.GetBucketVersioningOutput{
				Status: svcsdktypes.BucketVersioningStatusEnabled,
			}},
		}),
	}

	desired := newBucketResource("my-test-bucket")
	ko := desired.ko.DeepCopy()

	err := rm.addPutFieldsToSpec(context.Background(), desired, ko)

	require.Error(err)
	require.ErrorIs(err, requestPaymentErr)
}

// Test_addPutFieldsToSpec_ReadsAllFields verifies the happy path: when every
// Get* call succeeds, addPutFieldsToSpec reads Tagging, Versioning and Website
// (the fields that were being skipped by the bug) back into the spec. This
// guards against any future early-return regression of the same shape.
func Test_addPutFieldsToSpec_ReadsAllFields(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	rm := &resourceManager{
		sdkapi: newMockedSDKClient(map[string]opResult{
			"GetBucketRequestPayment": {output: &svcsdk.GetBucketRequestPaymentOutput{
				Payer: svcsdktypes.PayerRequester,
			}},
			"GetBucketTagging": {output: &svcsdk.GetBucketTaggingOutput{
				TagSet: []svcsdktypes.Tag{{Key: strPtr("owner"), Value: strPtr("ack")}},
			}},
			"GetBucketVersioning": {output: &svcsdk.GetBucketVersioningOutput{
				Status: svcsdktypes.BucketVersioningStatusEnabled,
			}},
		}),
	}

	desired := newBucketResource("my-test-bucket")
	ko := desired.ko.DeepCopy()

	err := rm.addPutFieldsToSpec(context.Background(), desired, ko)
	require.NoError(err)

	require.NotNil(ko.Spec.RequestPayment)
	assert.Equal(string(svcsdktypes.PayerRequester), *ko.Spec.RequestPayment.Payer)

	require.NotNil(ko.Spec.Versioning)
	assert.Equal(string(svcsdktypes.BucketVersioningStatusEnabled), *ko.Spec.Versioning.Status)

	require.NotNil(ko.Spec.Tagging)
	require.Len(ko.Spec.Tagging.TagSet, 1)
	assert.Equal("owner", *ko.Spec.Tagging.TagSet[0].Key)
	assert.Equal("ack", *ko.Spec.Tagging.TagSet[0].Value)
}

// Test_addPutFieldsToSpec_ObjectLockNotEnabledOnBucket is a regression test
// for aws-controllers-k8s/community#2965. When the bucket has no Object Lock
// configuration, the observed spec must report ObjectLockEnabledForBucket as
// false; it used to keep the stale desired `true` (ko is seeded from a copy
// of the desired resource), so adopted buckets never produced a delta and
// Object Lock was never enabled.
func Test_addPutFieldsToSpec_ObjectLockNotEnabledOnBucket(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	rm := &resourceManager{
		sdkapi: newMockedSDKClient(nil),
	}

	desired := newBucketResource("my-adopted-bucket")
	desired.ko.Spec.ObjectLockEnabledForBucket = boolPtr(true)
	ko := desired.ko.DeepCopy()

	err := rm.addPutFieldsToSpec(context.Background(), desired, ko)
	require.NoError(err)

	require.NotNil(ko.Spec.ObjectLockEnabledForBucket)
	assert.False(*ko.Spec.ObjectLockEnabledForBucket)
	assert.Nil(ko.Spec.ObjectLockConfiguration)
}

// Test_addPutFieldsToSpec_ObjectLockUnsetStaysNil verifies the reset only
// applies when the user set the field, since nil-vs-false would produce a
// permanent delta for buckets that never asked for Object Lock.
func Test_addPutFieldsToSpec_ObjectLockUnsetStaysNil(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	rm := &resourceManager{
		sdkapi: newMockedSDKClient(nil),
	}

	desired := newBucketResource("my-test-bucket")
	ko := desired.ko.DeepCopy()

	err := rm.addPutFieldsToSpec(context.Background(), desired, ko)
	require.NoError(err)

	assert.Nil(ko.Spec.ObjectLockEnabledForBucket)
}

// Test_addPutFieldsToSpec_ObjectLockEnabledOnBucket verifies a bucket that is
// already locked produces no delta.
func Test_addPutFieldsToSpec_ObjectLockEnabledOnBucket(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	rm := &resourceManager{
		sdkapi: newMockedSDKClient(map[string]opResult{
			"GetObjectLockConfiguration": {output: &svcsdk.GetObjectLockConfigurationOutput{
				ObjectLockConfiguration: &svcsdktypes.ObjectLockConfiguration{
					ObjectLockEnabled: svcsdktypes.ObjectLockEnabledEnabled,
				},
			}},
		}),
	}

	desired := newBucketResource("my-locked-bucket")
	desired.ko.Spec.ObjectLockEnabledForBucket = boolPtr(true)
	ko := desired.ko.DeepCopy()

	err := rm.addPutFieldsToSpec(context.Background(), desired, ko)
	require.NoError(err)

	require.NotNil(ko.Spec.ObjectLockEnabledForBucket)
	assert.True(*ko.Spec.ObjectLockEnabledForBucket)
}

func strPtr(s string) *string { return &s }

func boolPtr(b bool) *bool { return &b }
