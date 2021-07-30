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
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go/service/s3"
)

func (rm *resourceManager) createPutFields(
	ctx context.Context,
	r *resource,
) error {
	if err := rm.syncLogging(ctx, r); err != nil {
		return err
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

	if delta.DifferentAt("Spec.Logging") {
		if err := rm.syncLogging(ctx, desired); err != nil {
			return nil, err
		}
	}

	return &resource{ko}, nil
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
) (*svcsdk.PutBucketLoggingInput, error) {
	res := &svcsdk.PutBucketLoggingInput{}

	res.SetBucket(*r.ko.Spec.Name)
	if r.ko.Spec.Logging != nil {
		res.SetBucketLoggingStatus(rm.createBucketLoggingStatus(r))
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
	input, err := rm.newPutBucketLoggingPayload(r)
	if err != nil {
		return err
	}

	_, err = rm.sdkapi.PutBucketLogging(input)
	rm.metrics.RecordAPICall("UPDATED", "PutBucketLogging", err)
	if err != nil {
		return err
	}

	return nil
}
