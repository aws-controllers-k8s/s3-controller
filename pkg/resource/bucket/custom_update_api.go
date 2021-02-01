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
	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/s3"
)

// customUpdateBucket implements specialized logic for handling Bucket
// resource updates. The S3 API has 19 (!) separate API calls to update a
// Bucket, depending on the Bucket attribute that has changed. We currently
// support only those API calls that map to attributes of the Bucket that are
// settable on CreateBucket.:
//
// * PutBucketAccelerateConfiguration for when the
//   Bucket.AccelerateConfiguration struct changed (NOT SUPPORTED, since the
//   CreateBucket API call does not have an accelerate configuration element)
//   https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketAccelerateConfiguration.html
// * PutBucketAcl for when the Bucket's Access Control List attributes are
//   changed.
//   https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketAcl.html
// * PutBucketAnalyticsConfiguration for when the
//   Bucket.AnalyticsConfiguration struct changed (NOT SUPPORTED, since the
//   CreateBucket API call does not have an analytics configuration element)
//   https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketAnalyticsConfiguration.html
// * PutBucketCors for when the
//   Bucket.CORSConfiguration struct changed (NOT SUPPORTED, since the
//   CreateBucket API call does not have a CORS configuration element)
//   https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketCors.html
// * PutBucketEncryption for when the
//   Bucket.EncryptionConfiguration struct changed (NOT SUPPORTED, since the
//   CreateBucket API call does not have an encryption configuration element)
//   https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketEncryption.html
// * PutBucketIntelligentTieringConfiguration for when the
//   Bucket.IntelligentTieringConfiguration struct changed (NOT SUPPORTED,
//   since the CreateBucket API call does not have an intelligent tiering
//   configuration element)
//   https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketIntelligentTieringConfiguration.html
// * PutBucketInventoryConfiguration for when the Bucket.InventoryConfiguration
//   struct changed (NOT SUPPORTED, since the CreateBucket API call does not
//   have an inventory configuration element)
//   https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketInventoryConfiguration.html
// * PutBucketLifecycleConfiguration for when the Bucket.LifecycleConfiguration
//   struct changed (NOT SUPPORTED, since the CreateBucket API call does not
//   have a lifecycle configuration element)
//   https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketLifecycleConfiguration.html
// * PutBucketLogging for when the Bucket.LoggingConfiguration struct changed
//   (NOT SUPPORTED, since the CreateBucket API call does not have a logging
//   configuration element)
//   https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketLogging.html
// * PutBucketMetricsConfiguration for when the Bucket.MetricsConfiguration
//   struct changed (NOT SUPPORTED, since the CreateBucket API call does not
//   have a metrics configuration element)
//   https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketMetricsConfiguration.html
// * PutBucketNotificationConfiguration for when the
//   Bucket.NotificationConfiguration struct changed (NOT SUPPORTED, since the
//   CreateBucket API call does not have a notification configuration element)
//   https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketNotificationConfiguration.html
// * PutBucketOwnershipControls for when the Bucket.OwnershipControls struct
//   changed (NOT SUPPORTED, since the CreateBucket API call does not have a
//   ownership controls element)
//   https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketOwnershipControls.html
// * PutBucketPolicy for when the Bucket.Policy is changed.
//   https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketPolicy.htmlA
// * PutBucketReplication for when the Bucket.ReplicationConfiguration struct changed
//   (NOT SUPPORTED, since the CreateBucket API call does not have a replication
//   configuration element)
//   https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketReplication.html
// * PutBucketRequestPayment for when the Bucket.PaymentConfiguration struct
//   changed (NOT SUPPORTED, since the CreateBucket API call does not have a
//   payment configuration element)
//   https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketRequestPayment.html
// * PutBucketTagging for when the Bucket.Tags map  changed (NOT SUPPORTED,
//   since the CreateBucket API call does not have a tags element)
//   https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketTagging.html
// * PutBucketVersioning for when the Bucket.VersioningConfiguration struct
//   changed (NOT SUPPORTED, since the CreateBucket API call does not have a
//   versioning configuration element)
//   https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketVersioning.html
// * PutBucketWebsite for when the Bucket.WebsiteConfiguration struct
//   changed (NOT SUPPORTED, since the CreateBucket API call does not have a
//   website configuration element)
//   https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketWebsite.html
// * PutObjectLockConfiguration for the Bucket.ObjectLockEnabled bool is
//   changed.
//   https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutObjectLockConfiguration.html
func (rm *resourceManager) customUpdateBucket(
	ctx context.Context,
	desired *resource,
	latest *resource,
	delta *ackcompare.Delta,
) (*resource, error) {
	var err error
	var updated *resource
	updated = desired
	if delta.DifferentAt("ACL") ||
		delta.DifferentAt("GrantFullControll") ||
		delta.DifferentAt("GrantRead") ||
		delta.DifferentAt("GrantReadACP") ||
		delta.DifferentAt("GrantWrite") ||
		delta.DifferentAt("GrantWriteACP") {
		updated, err = rm.updateACL(ctx, updated)
		if err != nil {
			return nil, err
		}
	}
	if delta.DifferentAt("ObjectLockEnabledForBucket") {
		updated, err = rm.updateObjectLock(ctx, updated)
		if err != nil {
			return nil, err
		}
	}
	return updated, nil
}

// updateACL calls the PutBucketAcl S3 API call for a specific bucket
func (rm *resourceManager) updateACL(
	ctx context.Context,
	desired *resource,
) (*resource, error) {
	dspec := desired.ko.Spec
	input := &svcsdk.PutBucketAclInput{
		Bucket: aws.String(*dspec.Name),
	}
	if dspec.ACL != nil {
		input.SetACL(*dspec.ACL)
	}
	if dspec.GrantFullControl != nil {
		input.SetGrantFullControl(*dspec.GrantFullControl)
	}
	if dspec.GrantRead != nil {
		input.SetGrantRead(*dspec.GrantRead)
	}
	if dspec.GrantReadACP != nil {
		input.SetGrantReadACP(*dspec.GrantReadACP)
	}
	if dspec.GrantWrite != nil {
		input.SetGrantWrite(*dspec.GrantWrite)
	}
	if dspec.GrantWriteACP != nil {
		input.SetGrantWriteACP(*dspec.GrantWriteACP)
	}

	_, err := rm.sdkapi.PutBucketAclWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return desired, nil
}

// updateObjectLock calls the PutObjectLockConfiguration S3 API call for a
// specific bucket
func (rm *resourceManager) updateObjectLock(
	ctx context.Context,
	desired *resource,
) (*resource, error) {
	dspec := desired.ko.Spec
	input := &svcsdk.PutObjectLockConfigurationInput{
		Bucket: aws.String(*dspec.Name),
	}
	if dspec.ObjectLockEnabledForBucket != nil && *dspec.ObjectLockEnabledForBucket {
		olc := &svcsdk.ObjectLockConfiguration{}
		// Yep, this is NOT a typo. There is actually a const enum string
		// called ObjectLockEnabledEnabled
		olc.SetObjectLockEnabled(svcsdk.ObjectLockEnabledEnabled)
		input.SetObjectLockConfiguration(olc)
	}
	_, err := rm.sdkapi.PutObjectLockConfigurationWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	return desired, nil
}
