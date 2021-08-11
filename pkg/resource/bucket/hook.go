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

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackerr "github.com/aws-controllers-k8s/runtime/pkg/errors"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcapitypes "github.com/aws-controllers-k8s/s3-controller/apis/v1alpha1"
	svcsdk "github.com/aws/aws-sdk-go/service/s3"
)

var (
	DefaultAccelerationConfigurationStatus = "Suspended"
	DefaultRequestPayer                    = "BucketOwner"
)

func (rm *resourceManager) createPutFields(
	ctx context.Context,
	r *resource,
) error {
	if r.ko.Spec.Accelerate != nil {
		if err := rm.syncAccelerate(ctx, r); err != nil {
			return err
		}
	}
	if r.ko.Spec.CORS != nil {
		if err := rm.syncCORS(ctx, r); err != nil {
			return err
		}
	}
	if r.ko.Spec.Encryption != nil {
		if err := rm.syncEncryption(ctx, r); err != nil {
			return err
		}
	}
	if r.ko.Spec.Logging != nil {
		if err := rm.syncLogging(ctx, r); err != nil {
			return err
		}
	}
	if r.ko.Spec.OwnershipControls != nil {
		if err := rm.syncOwnershipControls(ctx, r); err != nil {
			return err
		}
	}
	if r.ko.Spec.RequestPayment != nil {
		if err := rm.syncRequestPayment(ctx, r); err != nil {
			return err
		}
	}
	if r.ko.Spec.Tagging != nil {
		if err := rm.syncTagging(ctx, r); err != nil {
			return err
		}
	}
	if r.ko.Spec.Website != nil {
		if err := rm.syncWebsite(ctx, r); err != nil {
			return err
		}
	}
	return nil
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

	// Merge in the information we read from the API call above to the copy of
	// the original Kubernetes object we passed to the function
	ko := desired.ko.DeepCopy()

	rm.setStatusDefaults(ko)

	if delta.DifferentAt("Spec.Accelerate") {
		if err := rm.syncAccelerate(ctx, desired); err != nil {
			return nil, err
		}
	}
	if delta.DifferentAt("Spec.CORS") {
		if err := rm.syncCORS(ctx, desired); err != nil {
			return nil, err
		}
	}
	if delta.DifferentAt("Spec.Encryption") {
		if err := rm.syncEncryption(ctx, desired); err != nil {
			return nil, err
		}
	}
	if delta.DifferentAt("Spec.Logging") {
		if err := rm.syncLogging(ctx, desired); err != nil {
			return nil, err
		}
	}
	if delta.DifferentAt("Spec.OwnershipControls") {
		if err := rm.syncOwnershipControls(ctx, desired); err != nil {
			return nil, err
		}
	}
	if delta.DifferentAt("Spec.RequestPayment") {
		if err := rm.syncRequestPayment(ctx, desired); err != nil {
			return nil, err
		}
	}
	if delta.DifferentAt("Spec.Tagging") {
		if err := rm.syncTagging(ctx, desired); err != nil {
			return nil, err
		}
	}
	if delta.DifferentAt("Spec.Website") {
		if err := rm.syncWebsite(ctx, desired); err != nil {
			return nil, err
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
	getAccelerateResponse, err := rm.sdkapi.GetBucketAccelerateConfigurationWithContext(ctx, rm.newGetBucketAcceleratePayload(r))
	if err != nil {
		return err
	}
	ko.Spec.Accelerate = rm.setResourceAccelerate(r, getAccelerateResponse)

	getCORSResponse, err := rm.sdkapi.GetBucketCorsWithContext(ctx, rm.newGetBucketCORSPayload(r))
	if err != nil {
		if awsErr, ok := ackerr.AWSError(err); ok && awsErr.Code() == "NoSuchCORSConfiguration" {
			getCORSResponse = &svcsdk.GetBucketCorsOutput{}
		} else {
			return err
		}
	}
	ko.Spec.CORS = rm.setResourceCORS(r, getCORSResponse)

	getEncryptionResponse, err := rm.sdkapi.GetBucketEncryptionWithContext(ctx, rm.newGetBucketEncryptionPayload(r))
	if err != nil {
		if awsErr, ok := ackerr.AWSError(err); ok && awsErr.Code() == "ServerSideEncryptionConfigurationNotFoundError" {
			getEncryptionResponse = &svcsdk.GetBucketEncryptionOutput{
				ServerSideEncryptionConfiguration: &svcsdk.ServerSideEncryptionConfiguration{},
			}
		} else {
			return err
		}
	}
	ko.Spec.Encryption = rm.setResourceEncryption(r, getEncryptionResponse)

	getLoggingResponse, err := rm.sdkapi.GetBucketLoggingWithContext(ctx, rm.newGetBucketLoggingPayload(r))
	if err != nil {
		return err
	}
	ko.Spec.Logging = rm.setResourceLogging(r, getLoggingResponse)

	getOwnershipControlsResponse, err := rm.sdkapi.GetBucketOwnershipControlsWithContext(ctx, rm.newGetBucketOwnershipControlsPayload(r))
	if err != nil {
		if awsErr, ok := ackerr.AWSError(err); ok && awsErr.Code() == "OwnershipControlsNotFoundError" {
			getOwnershipControlsResponse = &svcsdk.GetBucketOwnershipControlsOutput{
				OwnershipControls: &svcsdk.OwnershipControls{},
			}
		} else {
			return err
		}
	}
	ko.Spec.OwnershipControls = rm.setResourceOwnershipControls(r, getOwnershipControlsResponse)

	getRequestPaymentResponse, err := rm.sdkapi.GetBucketRequestPaymentWithContext(ctx, rm.newGetBucketRequestPaymentPayload(r))
	if err != nil {
		return nil
	}
	ko.Spec.RequestPayment = rm.setResourceRequestPayment(r, getRequestPaymentResponse)

	getTaggingResponse, err := rm.sdkapi.GetBucketTaggingWithContext(ctx, rm.newGetBucketTaggingPayload(r))
	if err != nil {
		if awsErr, ok := ackerr.AWSError(err); ok && awsErr.Code() == "NoSuchTagSet" {
			getTaggingResponse = &svcsdk.GetBucketTaggingOutput{}
		} else {
			return err
		}
	}
	ko.Spec.Tagging = rm.setResourceTagging(r, getTaggingResponse)

	getWebsiteResponse, err := rm.sdkapi.GetBucketWebsiteWithContext(ctx, rm.newGetBucketWebsitePayload(r))
	if err != nil {
		if awsErr, ok := ackerr.AWSError(err); ok && awsErr.Code() == "NoSuchWebsiteConfiguration" {
			getWebsiteResponse = &svcsdk.GetBucketWebsiteOutput{}
		} else {
			return err
		}
	}
	ko.Spec.Website = rm.setResourceWebsite(r, getWebsiteResponse)
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
		a.ko.Spec.Accelerate = &svcapitypes.AccelerateConfiguration{
			Status: &DefaultAccelerationConfigurationStatus,
		}
	}
	if a.ko.Spec.CORS == nil && b.ko.Spec.CORS != nil {
		a.ko.Spec.CORS = &svcapitypes.CORSConfiguration{}
	}
	if a.ko.Spec.Encryption == nil && b.ko.Spec.Encryption != nil {
		a.ko.Spec.Encryption = &svcapitypes.ServerSideEncryptionConfiguration{}
	}
	if a.ko.Spec.Logging == nil && b.ko.Spec.Logging != nil {
		a.ko.Spec.Logging = &svcapitypes.BucketLoggingStatus{}
	}
	if a.ko.Spec.OwnershipControls == nil && b.ko.Spec.OwnershipControls != nil {
		a.ko.Spec.OwnershipControls = &svcapitypes.OwnershipControls{}
	}
	if a.ko.Spec.RequestPayment == nil && b.ko.Spec.RequestPayment != nil {
		a.ko.Spec.RequestPayment = &svcapitypes.RequestPaymentConfiguration{
			Payer: &DefaultRequestPayer,
		}
	}
	if a.ko.Spec.Tagging == nil && b.ko.Spec.Tagging != nil {
		a.ko.Spec.Tagging = &svcapitypes.Tagging{}
	}
	if a.ko.Spec.Website == nil && b.ko.Spec.Website != nil {
		a.ko.Spec.Website = &svcapitypes.WebsiteConfiguration{}
	}
}

func (rm *resourceManager) newGetBucketAcceleratePayload(
	r *resource,
) *svcsdk.GetBucketAccelerateConfigurationInput {
	res := &svcsdk.GetBucketAccelerateConfigurationInput{}
	res.SetBucket(*r.ko.Spec.Name)
	return res
}

func (rm *resourceManager) newPutBucketAcceleratePayload(
	r *resource,
) *svcsdk.PutBucketAccelerateConfigurationInput {
	res := &svcsdk.PutBucketAccelerateConfigurationInput{}
	res.SetBucket(*r.ko.Spec.Name)
	if r.ko.Spec.Accelerate != nil {
		res.SetAccelerateConfiguration(rm.newAccelerateConfiguration(r))
	} else {
		res.SetAccelerateConfiguration(&svcsdk.AccelerateConfiguration{})
	}

	if res.AccelerateConfiguration.Status == nil {
		res.AccelerateConfiguration.SetStatus(DefaultAccelerationConfigurationStatus)
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

	_, err = rm.sdkapi.PutBucketAccelerateConfigurationWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketAccelerate", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) newGetBucketCORSPayload(
	r *resource,
) *svcsdk.GetBucketCorsInput {
	res := &svcsdk.GetBucketCorsInput{}
	res.SetBucket(*r.ko.Spec.Name)
	return res
}

func (rm *resourceManager) newPutBucketCORSPayload(
	r *resource,
) *svcsdk.PutBucketCorsInput {
	res := &svcsdk.PutBucketCorsInput{}
	res.SetBucket(*r.ko.Spec.Name)
	if r.ko.Spec.CORS != nil {
		res.SetCORSConfiguration(rm.newCORSConfiguration(r))
	} else {
		res.SetCORSConfiguration(&svcsdk.CORSConfiguration{})
	}
	return res
}

func (rm *resourceManager) syncCORS(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncCORS")
	defer exit(err)
	input := rm.newPutBucketCORSPayload(r)

	_, err = rm.sdkapi.PutBucketCorsWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketCors", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) newGetBucketEncryptionPayload(
	r *resource,
) *svcsdk.GetBucketEncryptionInput {
	res := &svcsdk.GetBucketEncryptionInput{}
	res.SetBucket(*r.ko.Spec.Name)
	return res
}

func (rm *resourceManager) newPutBucketEncryptionPayload(
	r *resource,
) *svcsdk.PutBucketEncryptionInput {
	res := &svcsdk.PutBucketEncryptionInput{}
	res.SetBucket(*r.ko.Spec.Name)
	if r.ko.Spec.Encryption != nil {
		res.SetServerSideEncryptionConfiguration(rm.newServerSideEncryptionConfiguration(r))
	} else {
		res.SetServerSideEncryptionConfiguration(&svcsdk.ServerSideEncryptionConfiguration{})
	}
	return res
}

func (rm *resourceManager) syncEncryption(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncEncryption")
	defer exit(err)
	input := rm.newPutBucketEncryptionPayload(r)

	_, err = rm.sdkapi.PutBucketEncryptionWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketEncryption", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) newGetBucketLoggingPayload(
	r *resource,
) *svcsdk.GetBucketLoggingInput {
	res := &svcsdk.GetBucketLoggingInput{}
	res.SetBucket(*r.ko.Spec.Name)
	return res
}

func (rm *resourceManager) newPutBucketLoggingPayload(
	r *resource,
) *svcsdk.PutBucketLoggingInput {
	res := &svcsdk.PutBucketLoggingInput{}
	res.SetBucket(*r.ko.Spec.Name)
	if r.ko.Spec.Logging != nil {
		res.SetBucketLoggingStatus(rm.newBucketLoggingStatus(r))
	} else {
		res.SetBucketLoggingStatus(&svcsdk.BucketLoggingStatus{})
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

	_, err = rm.sdkapi.PutBucketLoggingWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketLogging", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) newGetBucketOwnershipControlsPayload(
	r *resource,
) *svcsdk.GetBucketOwnershipControlsInput {
	res := &svcsdk.GetBucketOwnershipControlsInput{}
	res.SetBucket(*r.ko.Spec.Name)
	return res
}

func (rm *resourceManager) newPutBucketOwnershipControlsPayload(
	r *resource,
) *svcsdk.PutBucketOwnershipControlsInput {
	res := &svcsdk.PutBucketOwnershipControlsInput{}
	res.SetBucket(*r.ko.Spec.Name)
	if r.ko.Spec.OwnershipControls != nil {
		res.SetOwnershipControls(rm.newOwnershipControls(r))
	} else {
		res.SetOwnershipControls(&svcsdk.OwnershipControls{})
	}
	return res
}

func (rm *resourceManager) syncOwnershipControls(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncOwnershipControls")
	defer exit(err)
	input := rm.newPutBucketOwnershipControlsPayload(r)

	_, err = rm.sdkapi.PutBucketOwnershipControlsWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketOwnershipControls", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) newGetBucketRequestPaymentPayload(
	r *resource,
) *svcsdk.GetBucketRequestPaymentInput {
	res := &svcsdk.GetBucketRequestPaymentInput{}
	res.SetBucket(*r.ko.Spec.Name)
	return res
}

func (rm *resourceManager) newPutBucketRequestPaymentPayload(
	r *resource,
) *svcsdk.PutBucketRequestPaymentInput {
	res := &svcsdk.PutBucketRequestPaymentInput{}
	res.SetBucket(*r.ko.Spec.Name)
	if r.ko.Spec.RequestPayment != nil && r.ko.Spec.RequestPayment.Payer != nil {
		res.SetRequestPaymentConfiguration(rm.newRequestPaymentConfiguration(r))
	} else {
		res.SetRequestPaymentConfiguration(&svcsdk.RequestPaymentConfiguration{})
	}

	if res.RequestPaymentConfiguration.Payer == nil {
		res.RequestPaymentConfiguration.SetPayer(DefaultRequestPayer)
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

	_, err = rm.sdkapi.PutBucketRequestPaymentWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketRequestPayment", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) newGetBucketTaggingPayload(
	r *resource,
) *svcsdk.GetBucketTaggingInput {
	res := &svcsdk.GetBucketTaggingInput{}
	res.SetBucket(*r.ko.Spec.Name)
	return res
}

func (rm *resourceManager) newPutBucketTaggingPayload(
	r *resource,
) *svcsdk.PutBucketTaggingInput {
	res := &svcsdk.PutBucketTaggingInput{}
	res.SetBucket(*r.ko.Spec.Name)
	if r.ko.Spec.Tagging != nil {
		res.SetTagging(rm.newTagging(r))
	} else {
		res.SetTagging(&svcsdk.Tagging{})
	}
	return res
}

func (rm *resourceManager) syncTagging(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncTagging")
	defer exit(err)
	input := rm.newPutBucketTaggingPayload(r)

	_, err = rm.sdkapi.PutBucketTaggingWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketTagging", err)
	if err != nil {
		return err
	}

	return nil
}

func (rm *resourceManager) newGetBucketWebsitePayload(
	r *resource,
) *svcsdk.GetBucketWebsiteInput {
	res := &svcsdk.GetBucketWebsiteInput{}
	res.SetBucket(*r.ko.Spec.Name)
	return res
}

func (rm *resourceManager) newPutBucketWebsitePayload(
	r *resource,
) *svcsdk.PutBucketWebsiteInput {
	res := &svcsdk.PutBucketWebsiteInput{}
	res.SetBucket(*r.ko.Spec.Name)
	if r.ko.Spec.Website != nil {
		res.SetWebsiteConfiguration(rm.newWebsiteConfiguration(r))
	} else {
		res.SetWebsiteConfiguration(&svcsdk.WebsiteConfiguration{})
	}
	return res
}

func (rm *resourceManager) syncWebsite(
	ctx context.Context,
	r *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncWebsite")
	defer exit(err)
	input := rm.newPutBucketWebsitePayload(r)

	_, err = rm.sdkapi.PutBucketWebsiteWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketWebsite", err)
	if err != nil {
		return err
	}

	return nil
}
