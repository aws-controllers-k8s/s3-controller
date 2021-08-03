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

// Code generated by ack-generate. DO NOT EDIT.

package bucket

import (
	"context"
	"strings"

	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/s3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	svcapitypes "github.com/aws-controllers-k8s/s3-controller/apis/v1alpha1"
)

// Hack to avoid import errors during build...
var (
	_ = &metav1.Time{}
	_ = strings.ToLower("")
	_ = &aws.JSONValue{}
	_ = &svcsdk.S3{}
	_ = &svcapitypes.Bucket{}
	_ = ackv1alpha1.AWSAccountID("")
	_ = &ackerr.NotFound
)

// sdkFind returns SDK-specific information about a supplied resource
func (rm *resourceManager) sdkFind(
	ctx context.Context,
	r *resource,
) (latest *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkFind")
	defer exit(err)
	input, err := rm.newListRequestPayload(r)
	if err != nil {
		return nil, err
	}
	var resp *svcsdk.ListBucketsOutput
	resp, err = rm.sdkapi.ListBucketsWithContext(ctx, input)
	rm.metrics.RecordAPICall("READ_MANY", "ListBuckets", err)
	if err != nil {
		if awsErr, ok := ackerr.AWSError(err); ok && awsErr.Code() == "NoSuchBucket" {
			return nil, ackerr.NotFound
		}
		return nil, err
	}

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := r.ko.DeepCopy()

	found := false
	for _, elem := range resp.Buckets {
		if elem.Name != nil {
			if ko.Spec.Name != nil {
				if *elem.Name != *ko.Spec.Name {
					continue
				}
			}
			ko.Spec.Name = elem.Name
		} else {
			ko.Spec.Name = nil
		}
		found = true
		break
	}
	if !found {
		return nil, ackerr.NotFound
	}

	rm.setStatusDefaults(ko)
	if err := rm.addPutFieldsToSpec(ctx, r, ko); err != nil {
		return nil, err
	}
	return &resource{ko}, nil
}

// newListRequestPayload returns SDK-specific struct for the HTTP request
// payload of the List API call for the resource
func (rm *resourceManager) newListRequestPayload(
	r *resource,
) (*svcsdk.ListBucketsInput, error) {
	res := &svcsdk.ListBucketsInput{}

	return res, nil
}

// sdkCreate creates the supplied resource in the backend AWS service API and
// returns a copy of the resource with resource fields (in both Spec and
// Status) filled in with values from the CREATE API operation's Output shape.
func (rm *resourceManager) sdkCreate(
	ctx context.Context,
	desired *resource,
) (created *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkCreate")
	defer exit(err)
	input, err := rm.newCreateRequestPayload(ctx, desired)
	if err != nil {
		return nil, err
	}

	var resp *svcsdk.CreateBucketOutput
	_ = resp
	resp, err = rm.sdkapi.CreateBucketWithContext(ctx, input)
	rm.metrics.RecordAPICall("CREATE", "CreateBucket", err)
	if err != nil {
		return nil, err
	}
	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := desired.ko.DeepCopy()

	if resp.Location != nil {
		ko.Status.Location = resp.Location
	} else {
		ko.Status.Location = nil
	}

	rm.setStatusDefaults(ko)
	if err := rm.createPutFields(ctx, desired); err != nil {
		return nil, err
	}
	return &resource{ko}, nil
}

// newCreateRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Create API call for the resource
func (rm *resourceManager) newCreateRequestPayload(
	ctx context.Context,
	r *resource,
) (*svcsdk.CreateBucketInput, error) {
	res := &svcsdk.CreateBucketInput{}

	if r.ko.Spec.ACL != nil {
		res.SetACL(*r.ko.Spec.ACL)
	}
	if r.ko.Spec.Name != nil {
		res.SetBucket(*r.ko.Spec.Name)
	}
	if r.ko.Spec.CreateBucketConfiguration != nil {
		f2 := &svcsdk.CreateBucketConfiguration{}
		if r.ko.Spec.CreateBucketConfiguration.LocationConstraint != nil {
			f2.SetLocationConstraint(*r.ko.Spec.CreateBucketConfiguration.LocationConstraint)
		}
		res.SetCreateBucketConfiguration(f2)
	}
	if r.ko.Spec.GrantFullControl != nil {
		res.SetGrantFullControl(*r.ko.Spec.GrantFullControl)
	}
	if r.ko.Spec.GrantRead != nil {
		res.SetGrantRead(*r.ko.Spec.GrantRead)
	}
	if r.ko.Spec.GrantReadACP != nil {
		res.SetGrantReadACP(*r.ko.Spec.GrantReadACP)
	}
	if r.ko.Spec.GrantWrite != nil {
		res.SetGrantWrite(*r.ko.Spec.GrantWrite)
	}
	if r.ko.Spec.GrantWriteACP != nil {
		res.SetGrantWriteACP(*r.ko.Spec.GrantWriteACP)
	}
	if r.ko.Spec.ObjectLockEnabledForBucket != nil {
		res.SetObjectLockEnabledForBucket(*r.ko.Spec.ObjectLockEnabledForBucket)
	}

	return res, nil
}

// sdkUpdate patches the supplied resource in the backend AWS service API and
// returns a new resource with updated fields.
func (rm *resourceManager) sdkUpdate(
	ctx context.Context,
	desired *resource,
	latest *resource,
	delta *ackcompare.Delta,
) (*resource, error) {
	return rm.customUpdateBucket(ctx, desired, latest, delta)
}

// sdkDelete deletes the supplied resource in the backend AWS service API
func (rm *resourceManager) sdkDelete(
	ctx context.Context,
	r *resource,
) (latest *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sdkDelete")
	defer exit(err)
	input, err := rm.newDeleteRequestPayload(r)
	if err != nil {
		return nil, err
	}
	var resp *svcsdk.DeleteBucketOutput
	_ = resp
	resp, err = rm.sdkapi.DeleteBucketWithContext(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeleteBucket", err)
	return nil, err
}

// newDeleteRequestPayload returns an SDK-specific struct for the HTTP request
// payload of the Delete API call for the resource
func (rm *resourceManager) newDeleteRequestPayload(
	r *resource,
) (*svcsdk.DeleteBucketInput, error) {
	res := &svcsdk.DeleteBucketInput{}

	if r.ko.Spec.Name != nil {
		res.SetBucket(*r.ko.Spec.Name)
	}

	return res, nil
}

// setStatusDefaults sets default properties into supplied custom resource
func (rm *resourceManager) setStatusDefaults(
	ko *svcapitypes.Bucket,
) {
	if ko.Status.ACKResourceMetadata == nil {
		ko.Status.ACKResourceMetadata = &ackv1alpha1.ResourceMetadata{}
	}
	if ko.Status.ACKResourceMetadata.OwnerAccountID == nil {
		ko.Status.ACKResourceMetadata.OwnerAccountID = &rm.awsAccountID
	}
	if ko.Status.Conditions == nil {
		ko.Status.Conditions = []*ackv1alpha1.Condition{}
	}
}

// updateConditions returns updated resource, true; if conditions were updated
// else it returns nil, false
func (rm *resourceManager) updateConditions(
	r *resource,
	onSuccess bool,
	err error,
) (*resource, bool) {
	ko := r.ko.DeepCopy()
	rm.setStatusDefaults(ko)

	// Terminal condition
	var terminalCondition *ackv1alpha1.Condition = nil
	var recoverableCondition *ackv1alpha1.Condition = nil
	var syncCondition *ackv1alpha1.Condition = nil
	for _, condition := range ko.Status.Conditions {
		if condition.Type == ackv1alpha1.ConditionTypeTerminal {
			terminalCondition = condition
		}
		if condition.Type == ackv1alpha1.ConditionTypeRecoverable {
			recoverableCondition = condition
		}
		if condition.Type == ackv1alpha1.ConditionTypeResourceSynced {
			syncCondition = condition
		}
	}

	if rm.terminalAWSError(err) || err == ackerr.SecretTypeNotSupported || err == ackerr.SecretNotFound {
		if terminalCondition == nil {
			terminalCondition = &ackv1alpha1.Condition{
				Type: ackv1alpha1.ConditionTypeTerminal,
			}
			ko.Status.Conditions = append(ko.Status.Conditions, terminalCondition)
		}
		var errorMessage = ""
		if err == ackerr.SecretTypeNotSupported || err == ackerr.SecretNotFound {
			errorMessage = err.Error()
		} else {
			awsErr, _ := ackerr.AWSError(err)
			errorMessage = awsErr.Message()
		}
		terminalCondition.Status = corev1.ConditionTrue
		terminalCondition.Message = &errorMessage
	} else {
		// Clear the terminal condition if no longer present
		if terminalCondition != nil {
			terminalCondition.Status = corev1.ConditionFalse
			terminalCondition.Message = nil
		}
		// Handling Recoverable Conditions
		if err != nil {
			if recoverableCondition == nil {
				// Add a new Condition containing a non-terminal error
				recoverableCondition = &ackv1alpha1.Condition{
					Type: ackv1alpha1.ConditionTypeRecoverable,
				}
				ko.Status.Conditions = append(ko.Status.Conditions, recoverableCondition)
			}
			recoverableCondition.Status = corev1.ConditionTrue
			awsErr, _ := ackerr.AWSError(err)
			errorMessage := err.Error()
			if awsErr != nil {
				errorMessage = awsErr.Message()
			}
			recoverableCondition.Message = &errorMessage
		} else if recoverableCondition != nil {
			recoverableCondition.Status = corev1.ConditionFalse
			recoverableCondition.Message = nil
		}
	}
	// Required to avoid the "declared but not used" error in the default case
	_ = syncCondition
	if terminalCondition != nil || recoverableCondition != nil || syncCondition != nil {
		return &resource{ko}, true // updated
	}
	return nil, false // not updated
}

// terminalAWSError returns awserr, true; if the supplied error is an aws Error type
// and if the exception indicates that it is a Terminal exception
// 'Terminal' exception are specified in generator configuration
func (rm *resourceManager) terminalAWSError(err error) bool {
	// No terminal_errors specified for this resource in generator config
	return false
}

// newAccelerateConfiguration returns a AccelerateConfiguration object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newAccelerateConfiguration(
	r *resource,
) *svcsdk.AccelerateConfiguration {
	res := &svcsdk.AccelerateConfiguration{}

	if r.ko.Spec.AccelerateConfiguration.Status != nil {
		res.SetStatus(*r.ko.Spec.AccelerateConfiguration.Status)
	}

	return res
}

// setResourceAccelerateConfiguration sets the `AccelerateConfiguration` spec field
// given the output of a `GetBucketAccelerateConfiguration` operation.
func (rm *resourceManager) setResourceAccelerateConfiguration(
	r *resource,
	resp *svcsdk.GetBucketAccelerateConfigurationOutput,
) *svcapitypes.AccelerateConfiguration {
	res := &svcapitypes.AccelerateConfiguration{}
	if resp.Status != nil {
		res.Status = resp.Status
	}

	return res
}

// newCORSConfiguration returns a CORSConfiguration object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newCORSConfiguration(
	r *resource,
) *svcsdk.CORSConfiguration {
	res := &svcsdk.CORSConfiguration{}

	if r.ko.Spec.CORS.CORSRules != nil {
		resf0 := []*svcsdk.CORSRule{}
		for _, resf0iter := range r.ko.Spec.CORS.CORSRules {
			resf0elem := &svcsdk.CORSRule{}
			if resf0iter.AllowedHeaders != nil {
				resf0elemf0 := []*string{}
				for _, resf0elemf0iter := range resf0iter.AllowedHeaders {
					var resf0elemf0elem string
					resf0elemf0elem = *resf0elemf0iter
					resf0elemf0 = append(resf0elemf0, &resf0elemf0elem)
				}
				resf0elem.SetAllowedHeaders(resf0elemf0)
			}
			if resf0iter.AllowedMethods != nil {
				resf0elemf1 := []*string{}
				for _, resf0elemf1iter := range resf0iter.AllowedMethods {
					var resf0elemf1elem string
					resf0elemf1elem = *resf0elemf1iter
					resf0elemf1 = append(resf0elemf1, &resf0elemf1elem)
				}
				resf0elem.SetAllowedMethods(resf0elemf1)
			}
			if resf0iter.AllowedOrigins != nil {
				resf0elemf2 := []*string{}
				for _, resf0elemf2iter := range resf0iter.AllowedOrigins {
					var resf0elemf2elem string
					resf0elemf2elem = *resf0elemf2iter
					resf0elemf2 = append(resf0elemf2, &resf0elemf2elem)
				}
				resf0elem.SetAllowedOrigins(resf0elemf2)
			}
			if resf0iter.ExposeHeaders != nil {
				resf0elemf3 := []*string{}
				for _, resf0elemf3iter := range resf0iter.ExposeHeaders {
					var resf0elemf3elem string
					resf0elemf3elem = *resf0elemf3iter
					resf0elemf3 = append(resf0elemf3, &resf0elemf3elem)
				}
				resf0elem.SetExposeHeaders(resf0elemf3)
			}
			if resf0iter.MaxAgeSeconds != nil {
				resf0elem.SetMaxAgeSeconds(*resf0iter.MaxAgeSeconds)
			}
			resf0 = append(resf0, resf0elem)
		}
		res.SetCORSRules(resf0)
	}

	return res
}

// setResourceCORS sets the `CORS` spec field
// given the output of a `GetBucketCors` operation.
func (rm *resourceManager) setResourceCORS(
	r *resource,
	resp *svcsdk.GetBucketCorsOutput,
) *svcapitypes.CORSConfiguration {
	res := &svcapitypes.CORSConfiguration{}
	if resp.CORSRules != nil {
		resf0 := []*svcapitypes.CORSRule{}
		for _, resf0iter := range resp.CORSRules {
			resf0elem := &svcapitypes.CORSRule{}
			if resf0iter.AllowedHeaders != nil {
				resf0elemf0 := []*string{}
				for _, resf0elemf0iter := range resf0iter.AllowedHeaders {
					var resf0elemf0elem string
					resf0elemf0elem = *resf0elemf0iter
					resf0elemf0 = append(resf0elemf0, &resf0elemf0elem)
				}
				resf0elem.AllowedHeaders = resf0elemf0
			}
			if resf0iter.AllowedMethods != nil {
				resf0elemf1 := []*string{}
				for _, resf0elemf1iter := range resf0iter.AllowedMethods {
					var resf0elemf1elem string
					resf0elemf1elem = *resf0elemf1iter
					resf0elemf1 = append(resf0elemf1, &resf0elemf1elem)
				}
				resf0elem.AllowedMethods = resf0elemf1
			}
			if resf0iter.AllowedOrigins != nil {
				resf0elemf2 := []*string{}
				for _, resf0elemf2iter := range resf0iter.AllowedOrigins {
					var resf0elemf2elem string
					resf0elemf2elem = *resf0elemf2iter
					resf0elemf2 = append(resf0elemf2, &resf0elemf2elem)
				}
				resf0elem.AllowedOrigins = resf0elemf2
			}
			if resf0iter.ExposeHeaders != nil {
				resf0elemf3 := []*string{}
				for _, resf0elemf3iter := range resf0iter.ExposeHeaders {
					var resf0elemf3elem string
					resf0elemf3elem = *resf0elemf3iter
					resf0elemf3 = append(resf0elemf3, &resf0elemf3elem)
				}
				resf0elem.ExposeHeaders = resf0elemf3
			}
			if resf0iter.MaxAgeSeconds != nil {
				resf0elem.MaxAgeSeconds = resf0iter.MaxAgeSeconds
			}
			resf0 = append(resf0, resf0elem)
		}
		res.CORSRules = resf0
	}

	return res
}

// newServerSideEncryptionConfiguration returns a ServerSideEncryptionConfiguration object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newServerSideEncryptionConfiguration(
	r *resource,
) *svcsdk.ServerSideEncryptionConfiguration {
	res := &svcsdk.ServerSideEncryptionConfiguration{}

	if r.ko.Spec.Encryption.Rules != nil {
		resf0 := []*svcsdk.ServerSideEncryptionRule{}
		for _, resf0iter := range r.ko.Spec.Encryption.Rules {
			resf0elem := &svcsdk.ServerSideEncryptionRule{}
			if resf0iter.ApplyServerSideEncryptionByDefault != nil {
				resf0elemf0 := &svcsdk.ServerSideEncryptionByDefault{}
				if resf0iter.ApplyServerSideEncryptionByDefault.KMSMasterKeyID != nil {
					resf0elemf0.SetKMSMasterKeyID(*resf0iter.ApplyServerSideEncryptionByDefault.KMSMasterKeyID)
				}
				if resf0iter.ApplyServerSideEncryptionByDefault.SSEAlgorithm != nil {
					resf0elemf0.SetSSEAlgorithm(*resf0iter.ApplyServerSideEncryptionByDefault.SSEAlgorithm)
				}
				resf0elem.SetApplyServerSideEncryptionByDefault(resf0elemf0)
			}
			if resf0iter.BucketKeyEnabled != nil {
				resf0elem.SetBucketKeyEnabled(*resf0iter.BucketKeyEnabled)
			}
			resf0 = append(resf0, resf0elem)
		}
		res.SetRules(resf0)
	}

	return res
}

// setResourceEncryption sets the `Encryption` spec field
// given the output of a `GetBucketEncryption` operation.
func (rm *resourceManager) setResourceEncryption(
	r *resource,
	resp *svcsdk.GetBucketEncryptionOutput,
) *svcapitypes.ServerSideEncryptionConfiguration {
	res := &svcapitypes.ServerSideEncryptionConfiguration{}
	if resp.ServerSideEncryptionConfiguration.Rules != nil {
		resf0 := []*svcapitypes.ServerSideEncryptionRule{}
		for _, resf0iter := range resp.ServerSideEncryptionConfiguration.Rules {
			resf0elem := &svcapitypes.ServerSideEncryptionRule{}
			if resf0iter.ApplyServerSideEncryptionByDefault != nil {
				resf0elemf0 := &svcapitypes.ServerSideEncryptionByDefault{}
				if resf0iter.ApplyServerSideEncryptionByDefault.KMSMasterKeyID != nil {
					resf0elemf0.KMSMasterKeyID = resf0iter.ApplyServerSideEncryptionByDefault.KMSMasterKeyID
				}
				if resf0iter.ApplyServerSideEncryptionByDefault.SSEAlgorithm != nil {
					resf0elemf0.SSEAlgorithm = resf0iter.ApplyServerSideEncryptionByDefault.SSEAlgorithm
				}
				resf0elem.ApplyServerSideEncryptionByDefault = resf0elemf0
			}
			if resf0iter.BucketKeyEnabled != nil {
				resf0elem.BucketKeyEnabled = resf0iter.BucketKeyEnabled
			}
			resf0 = append(resf0, resf0elem)
		}
		res.Rules = resf0
	}

	return res
}

// newBucketLoggingStatus returns a BucketLoggingStatus object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newBucketLoggingStatus(
	r *resource,
) *svcsdk.BucketLoggingStatus {
	res := &svcsdk.BucketLoggingStatus{}

	if r.ko.Spec.Logging.LoggingEnabled != nil {
		resf0 := &svcsdk.LoggingEnabled{}
		if r.ko.Spec.Logging.LoggingEnabled.TargetBucket != nil {
			resf0.SetTargetBucket(*r.ko.Spec.Logging.LoggingEnabled.TargetBucket)
		}
		if r.ko.Spec.Logging.LoggingEnabled.TargetGrants != nil {
			resf0f1 := []*svcsdk.TargetGrant{}
			for _, resf0f1iter := range r.ko.Spec.Logging.LoggingEnabled.TargetGrants {
				resf0f1elem := &svcsdk.TargetGrant{}
				if resf0f1iter.Grantee != nil {
					resf0f1elemf0 := &svcsdk.Grantee{}
					if resf0f1iter.Grantee.DisplayName != nil {
						resf0f1elemf0.SetDisplayName(*resf0f1iter.Grantee.DisplayName)
					}
					if resf0f1iter.Grantee.EmailAddress != nil {
						resf0f1elemf0.SetEmailAddress(*resf0f1iter.Grantee.EmailAddress)
					}
					if resf0f1iter.Grantee.ID != nil {
						resf0f1elemf0.SetID(*resf0f1iter.Grantee.ID)
					}
					if resf0f1iter.Grantee.Type != nil {
						resf0f1elemf0.SetType(*resf0f1iter.Grantee.Type)
					}
					if resf0f1iter.Grantee.URI != nil {
						resf0f1elemf0.SetURI(*resf0f1iter.Grantee.URI)
					}
					resf0f1elem.SetGrantee(resf0f1elemf0)
				}
				if resf0f1iter.Permission != nil {
					resf0f1elem.SetPermission(*resf0f1iter.Permission)
				}
				resf0f1 = append(resf0f1, resf0f1elem)
			}
			resf0.SetTargetGrants(resf0f1)
		}
		if r.ko.Spec.Logging.LoggingEnabled.TargetPrefix != nil {
			resf0.SetTargetPrefix(*r.ko.Spec.Logging.LoggingEnabled.TargetPrefix)
		}
		res.SetLoggingEnabled(resf0)
	}

	return res
}

// setResourceLogging sets the `Logging` spec field
// given the output of a `GetBucketLogging` operation.
func (rm *resourceManager) setResourceLogging(
	r *resource,
	resp *svcsdk.GetBucketLoggingOutput,
) *svcapitypes.BucketLoggingStatus {
	res := &svcapitypes.BucketLoggingStatus{}
	if resp.LoggingEnabled != nil {
		resf0 := &svcapitypes.LoggingEnabled{}
		if resp.LoggingEnabled.TargetBucket != nil {
			resf0.TargetBucket = resp.LoggingEnabled.TargetBucket
		}
		if resp.LoggingEnabled.TargetGrants != nil {
			resf0f1 := []*svcapitypes.TargetGrant{}
			for _, resf0f1iter := range resp.LoggingEnabled.TargetGrants {
				resf0f1elem := &svcapitypes.TargetGrant{}
				if resf0f1iter.Grantee != nil {
					resf0f1elemf0 := &svcapitypes.Grantee{}
					if resf0f1iter.Grantee.DisplayName != nil {
						resf0f1elemf0.DisplayName = resf0f1iter.Grantee.DisplayName
					}
					if resf0f1iter.Grantee.EmailAddress != nil {
						resf0f1elemf0.EmailAddress = resf0f1iter.Grantee.EmailAddress
					}
					if resf0f1iter.Grantee.ID != nil {
						resf0f1elemf0.ID = resf0f1iter.Grantee.ID
					}
					if resf0f1iter.Grantee.Type != nil {
						resf0f1elemf0.Type = resf0f1iter.Grantee.Type
					}
					if resf0f1iter.Grantee.URI != nil {
						resf0f1elemf0.URI = resf0f1iter.Grantee.URI
					}
					resf0f1elem.Grantee = resf0f1elemf0
				}
				if resf0f1iter.Permission != nil {
					resf0f1elem.Permission = resf0f1iter.Permission
				}
				resf0f1 = append(resf0f1, resf0f1elem)
			}
			resf0.TargetGrants = resf0f1
		}
		if resp.LoggingEnabled.TargetPrefix != nil {
			resf0.TargetPrefix = resp.LoggingEnabled.TargetPrefix
		}
		res.LoggingEnabled = resf0
	}

	return res
}

// newOwnershipControls returns a OwnershipControls object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newOwnershipControls(
	r *resource,
) *svcsdk.OwnershipControls {
	res := &svcsdk.OwnershipControls{}

	if r.ko.Spec.OwnershipControls.Rules != nil {
		resf0 := []*svcsdk.OwnershipControlsRule{}
		for _, resf0iter := range r.ko.Spec.OwnershipControls.Rules {
			resf0elem := &svcsdk.OwnershipControlsRule{}
			if resf0iter.ObjectOwnership != nil {
				resf0elem.SetObjectOwnership(*resf0iter.ObjectOwnership)
			}
			resf0 = append(resf0, resf0elem)
		}
		res.SetRules(resf0)
	}

	return res
}

// setResourceOwnershipControls sets the `OwnershipControls` spec field
// given the output of a `GetBucketOwnershipControls` operation.
func (rm *resourceManager) setResourceOwnershipControls(
	r *resource,
	resp *svcsdk.GetBucketOwnershipControlsOutput,
) *svcapitypes.OwnershipControls {
	res := &svcapitypes.OwnershipControls{}
	if resp.OwnershipControls.Rules != nil {
		resf0 := []*svcapitypes.OwnershipControlsRule{}
		for _, resf0iter := range resp.OwnershipControls.Rules {
			resf0elem := &svcapitypes.OwnershipControlsRule{}
			if resf0iter.ObjectOwnership != nil {
				resf0elem.ObjectOwnership = resf0iter.ObjectOwnership
			}
			resf0 = append(resf0, resf0elem)
		}
		res.Rules = resf0
	}

	return res
}

// newRequestPaymentConfiguration returns a RequestPaymentConfiguration object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newRequestPaymentConfiguration(
	r *resource,
) *svcsdk.RequestPaymentConfiguration {
	res := &svcsdk.RequestPaymentConfiguration{}

	if r.ko.Spec.RequestPayment.Payer != nil {
		res.SetPayer(*r.ko.Spec.RequestPayment.Payer)
	}

	return res
}

// setResourceRequestPayment sets the `RequestPayment` spec field
// given the output of a `GetBucketRequestPayment` operation.
func (rm *resourceManager) setResourceRequestPayment(
	r *resource,
	resp *svcsdk.GetBucketRequestPaymentOutput,
) *svcapitypes.RequestPaymentConfiguration {
	res := &svcapitypes.RequestPaymentConfiguration{}
	if resp.Payer != nil {
		res.Payer = resp.Payer
	}

	return res
}

// newTagging returns a Tagging object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newTagging(
	r *resource,
) *svcsdk.Tagging {
	res := &svcsdk.Tagging{}

	if r.ko.Spec.Tagging.TagSet != nil {
		resf0 := []*svcsdk.Tag{}
		for _, resf0iter := range r.ko.Spec.Tagging.TagSet {
			resf0elem := &svcsdk.Tag{}
			if resf0iter.Key != nil {
				resf0elem.SetKey(*resf0iter.Key)
			}
			if resf0iter.Value != nil {
				resf0elem.SetValue(*resf0iter.Value)
			}
			resf0 = append(resf0, resf0elem)
		}
		res.SetTagSet(resf0)
	}

	return res
}

// setResourceTagging sets the `Tagging` spec field
// given the output of a `GetBucketTagging` operation.
func (rm *resourceManager) setResourceTagging(
	r *resource,
	resp *svcsdk.GetBucketTaggingOutput,
) *svcapitypes.Tagging {
	res := &svcapitypes.Tagging{}
	if resp.TagSet != nil {
		resf0 := []*svcapitypes.Tag{}
		for _, resf0iter := range resp.TagSet {
			resf0elem := &svcapitypes.Tag{}
			if resf0iter.Key != nil {
				resf0elem.Key = resf0iter.Key
			}
			if resf0iter.Value != nil {
				resf0elem.Value = resf0iter.Value
			}
			resf0 = append(resf0, resf0elem)
		}
		res.TagSet = resf0
	}

	return res
}

// newWebsiteConfiguration returns a WebsiteConfiguration object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newWebsiteConfiguration(
	r *resource,
) *svcsdk.WebsiteConfiguration {
	res := &svcsdk.WebsiteConfiguration{}

	if r.ko.Spec.Website.ErrorDocument != nil {
		resf0 := &svcsdk.ErrorDocument{}
		if r.ko.Spec.Website.ErrorDocument.Key != nil {
			resf0.SetKey(*r.ko.Spec.Website.ErrorDocument.Key)
		}
		res.SetErrorDocument(resf0)
	}
	if r.ko.Spec.Website.IndexDocument != nil {
		resf1 := &svcsdk.IndexDocument{}
		if r.ko.Spec.Website.IndexDocument.Suffix != nil {
			resf1.SetSuffix(*r.ko.Spec.Website.IndexDocument.Suffix)
		}
		res.SetIndexDocument(resf1)
	}
	if r.ko.Spec.Website.RedirectAllRequestsTo != nil {
		resf2 := &svcsdk.RedirectAllRequestsTo{}
		if r.ko.Spec.Website.RedirectAllRequestsTo.HostName != nil {
			resf2.SetHostName(*r.ko.Spec.Website.RedirectAllRequestsTo.HostName)
		}
		if r.ko.Spec.Website.RedirectAllRequestsTo.Protocol != nil {
			resf2.SetProtocol(*r.ko.Spec.Website.RedirectAllRequestsTo.Protocol)
		}
		res.SetRedirectAllRequestsTo(resf2)
	}
	if r.ko.Spec.Website.RoutingRules != nil {
		resf3 := []*svcsdk.RoutingRule{}
		for _, resf3iter := range r.ko.Spec.Website.RoutingRules {
			resf3elem := &svcsdk.RoutingRule{}
			if resf3iter.Condition != nil {
				resf3elemf0 := &svcsdk.Condition{}
				if resf3iter.Condition.HTTPErrorCodeReturnedEquals != nil {
					resf3elemf0.SetHttpErrorCodeReturnedEquals(*resf3iter.Condition.HTTPErrorCodeReturnedEquals)
				}
				if resf3iter.Condition.KeyPrefixEquals != nil {
					resf3elemf0.SetKeyPrefixEquals(*resf3iter.Condition.KeyPrefixEquals)
				}
				resf3elem.SetCondition(resf3elemf0)
			}
			if resf3iter.Redirect != nil {
				resf3elemf1 := &svcsdk.Redirect{}
				if resf3iter.Redirect.HostName != nil {
					resf3elemf1.SetHostName(*resf3iter.Redirect.HostName)
				}
				if resf3iter.Redirect.HTTPRedirectCode != nil {
					resf3elemf1.SetHttpRedirectCode(*resf3iter.Redirect.HTTPRedirectCode)
				}
				if resf3iter.Redirect.Protocol != nil {
					resf3elemf1.SetProtocol(*resf3iter.Redirect.Protocol)
				}
				if resf3iter.Redirect.ReplaceKeyPrefixWith != nil {
					resf3elemf1.SetReplaceKeyPrefixWith(*resf3iter.Redirect.ReplaceKeyPrefixWith)
				}
				if resf3iter.Redirect.ReplaceKeyWith != nil {
					resf3elemf1.SetReplaceKeyWith(*resf3iter.Redirect.ReplaceKeyWith)
				}
				resf3elem.SetRedirect(resf3elemf1)
			}
			resf3 = append(resf3, resf3elem)
		}
		res.SetRoutingRules(resf3)
	}

	return res
}

// setResourceWebsite sets the `Website` spec field
// given the output of a `GetBucketWebsite` operation.
func (rm *resourceManager) setResourceWebsite(
	r *resource,
	resp *svcsdk.GetBucketWebsiteOutput,
) *svcapitypes.WebsiteConfiguration {
	res := &svcapitypes.WebsiteConfiguration{}
	if resp.ErrorDocument != nil {
		resf0 := &svcapitypes.ErrorDocument{}
		if resp.ErrorDocument.Key != nil {
			resf0.Key = resp.ErrorDocument.Key
		}
		res.ErrorDocument = resf0
	}
	if resp.IndexDocument != nil {
		resf1 := &svcapitypes.IndexDocument{}
		if resp.IndexDocument.Suffix != nil {
			resf1.Suffix = resp.IndexDocument.Suffix
		}
		res.IndexDocument = resf1
	}
	if resp.RedirectAllRequestsTo != nil {
		resf2 := &svcapitypes.RedirectAllRequestsTo{}
		if resp.RedirectAllRequestsTo.HostName != nil {
			resf2.HostName = resp.RedirectAllRequestsTo.HostName
		}
		if resp.RedirectAllRequestsTo.Protocol != nil {
			resf2.Protocol = resp.RedirectAllRequestsTo.Protocol
		}
		res.RedirectAllRequestsTo = resf2
	}
	if resp.RoutingRules != nil {
		resf3 := []*svcapitypes.RoutingRule{}
		for _, resf3iter := range resp.RoutingRules {
			resf3elem := &svcapitypes.RoutingRule{}
			if resf3iter.Condition != nil {
				resf3elemf0 := &svcapitypes.Condition{}
				if resf3iter.Condition.HttpErrorCodeReturnedEquals != nil {
					resf3elemf0.HTTPErrorCodeReturnedEquals = resf3iter.Condition.HttpErrorCodeReturnedEquals
				}
				if resf3iter.Condition.KeyPrefixEquals != nil {
					resf3elemf0.KeyPrefixEquals = resf3iter.Condition.KeyPrefixEquals
				}
				resf3elem.Condition = resf3elemf0
			}
			if resf3iter.Redirect != nil {
				resf3elemf1 := &svcapitypes.Redirect{}
				if resf3iter.Redirect.HostName != nil {
					resf3elemf1.HostName = resf3iter.Redirect.HostName
				}
				if resf3iter.Redirect.HttpRedirectCode != nil {
					resf3elemf1.HTTPRedirectCode = resf3iter.Redirect.HttpRedirectCode
				}
				if resf3iter.Redirect.Protocol != nil {
					resf3elemf1.Protocol = resf3iter.Redirect.Protocol
				}
				if resf3iter.Redirect.ReplaceKeyPrefixWith != nil {
					resf3elemf1.ReplaceKeyPrefixWith = resf3iter.Redirect.ReplaceKeyPrefixWith
				}
				if resf3iter.Redirect.ReplaceKeyWith != nil {
					resf3elemf1.ReplaceKeyWith = resf3iter.Redirect.ReplaceKeyWith
				}
				resf3elem.Redirect = resf3elemf1
			}
			resf3 = append(resf3, resf3elem)
		}
		res.RoutingRules = resf3
	}

	return res
}
