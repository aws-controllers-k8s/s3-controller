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

func (rm *resourceManager) syncPutFields(
	ctx context.Context,
	r *resource,
) (synced *resource, err error) {
	// Copy resource for later patching
	synced = &resource{r.ko.DeepCopy()}

	diff, err := rm.diffLogging(ctx, synced)
	if err != nil {
		return nil, err
	}

	if len(diff.Differences) > 0 {
		rm.syncLogging(ctx, synced)
	}

	return synced, nil
}

func (rm *resourceManager) diffLogging(
	ctx context.Context,
	r *resource,
) (*ackcompare.Delta, error) {
	desired := r.ko.Spec.Logging

	input := &svcsdk.GetBucketLoggingInput{
		Bucket: r.ko.Spec.Name,
	}
	latest, err := rm.sdkapi.GetBucketLoggingWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	delta := ackcompare.NewDelta()

	if ackcompare.HasNilDifference(desired.LoggingEnabled, latest.LoggingEnabled) {
		delta.Add("Spec.Logging.LoggingEnabled", desired.LoggingEnabled, latest.LoggingEnabled)
	} else {
		if ackcompare.HasNilDifference(desired.LoggingEnabled.TargetBucket, latest.LoggingEnabled.TargetBucket) {
			delta.Add("Spec.Logging.LoggingEnabled.TargetBucket", desired.LoggingEnabled.TargetBucket, latest.LoggingEnabled.TargetBucket)
		} else {
			if *desired.LoggingEnabled.TargetBucket != *latest.LoggingEnabled.TargetBucket {
				delta.Add("Spec.Logging.LoggingEnabled.TargetBucket", desired.LoggingEnabled.TargetBucket, latest.LoggingEnabled.TargetBucket)
			}
		}

		if ackcompare.HasNilDifference(desired.LoggingEnabled.TargetPrefix, latest.LoggingEnabled.TargetPrefix) {
			delta.Add("Spec.Logging.LoggingEnabled.TargetPrefix", desired.LoggingEnabled.TargetPrefix, latest.LoggingEnabled.TargetPrefix)
		} else {
			if *desired.LoggingEnabled.TargetPrefix != *latest.LoggingEnabled.TargetPrefix {
				delta.Add("Spec.Logging.LoggingEnabled.TargetPrefix", desired.LoggingEnabled.TargetPrefix, latest.LoggingEnabled.TargetPrefix)
			}
		}
	}

	// TODO(RedbackThomson): Diff LoggingEnabled.TargetGrants
	return delta, nil
}

func (rm *resourceManager) newPutBucketLoggingPayload(
	r *resource,
) (*svcsdk.PutBucketLoggingInput, error) {
	res := &svcsdk.PutBucketLoggingInput{}
	logging := r.ko.Spec.Logging

	res.SetBucket(*r.ko.Spec.Name)

	if logging != nil {
		loggingStatus := &svcsdk.BucketLoggingStatus{}

		if logging.LoggingEnabled != nil {
			loggingEnabled := &svcsdk.LoggingEnabled{}

			if logging.LoggingEnabled.TargetBucket != nil {
				loggingEnabled.SetTargetBucket(*logging.LoggingEnabled.TargetBucket)
			}
			if logging.LoggingEnabled.TargetPrefix != nil {
				loggingEnabled.SetTargetPrefix(*logging.LoggingEnabled.TargetPrefix)
			}

			grants := []*svcsdk.TargetGrant{}
			for _, grant := range logging.LoggingEnabled.TargetGrants {
				newGrant := &svcsdk.TargetGrant{}

				if grant.Permission != nil {
					newGrant.SetPermission(*grant.Permission)
				}

				if grant.Grantee != nil {
					newGrantee := &svcsdk.Grantee{}

					if grant.Grantee.DisplayName != nil {
						newGrantee.SetDisplayName(*grant.Grantee.DisplayName)
					}

					if grant.Grantee.EmailAddress != nil {
						newGrantee.SetEmailAddress(*grant.Grantee.EmailAddress)
					}

					if grant.Grantee.ID != nil {
						newGrantee.SetID(*grant.Grantee.ID)
					}

					if grant.Grantee.Type != nil {
						newGrantee.SetType(*grant.Grantee.Type)
					}

					if grant.Grantee.URI != nil {
						newGrantee.SetURI(*grant.Grantee.URI)
					}
				}

				grants = append(grants, newGrant)
			}
			if len(grants) > 0 {
				loggingEnabled.SetTargetGrants(grants)
			}

			loggingStatus.SetLoggingEnabled(loggingEnabled)
		}
		res.SetBucketLoggingStatus(loggingStatus)
	}

	return res, nil
}

func (rm *resourceManager) syncLogging(
	ctx context.Context,
	r *resource,
) (created *resource, err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncLogging")
	defer exit(err)
	input, err := rm.newPutBucketLoggingPayload(r)
	if err != nil {
		return nil, err
	}

	ko := r.ko.DeepCopy()

	_, err = rm.sdkapi.PutBucketLogging(input)
	rm.metrics.RecordAPICall("UPDATED", "PutBucketLogging", err)
	if err != nil {
		return nil, err
	}

	return &resource{ko}, nil
}
