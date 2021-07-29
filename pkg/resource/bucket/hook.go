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
