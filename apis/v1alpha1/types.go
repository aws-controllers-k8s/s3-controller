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

package v1alpha1

import (
	ackv1alpha1 "github.com/aws-controllers-k8s/runtime/apis/core/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Hack to avoid import errors during build...
var (
	_ = &metav1.Time{}
	_ = &aws.JSONValue{}
	_ = ackv1alpha1.AWSAccountID("")
)

// Configures the transfer acceleration state for an Amazon S3 bucket. For more
// information, see Amazon S3 Transfer Acceleration (https://docs.aws.amazon.com/AmazonS3/latest/dev/transfer-acceleration.html)
// in the Amazon Simple Storage Service Developer Guide.
type AccelerateConfiguration struct {
	Status *string `json:"status,omitempty"`
}

// Contains the elements that set the ACL permissions for an object per grantee.
type AccessControlPolicy struct {
	// Container for the owner's display name and ID.
	Owner *Owner `json:"owner,omitempty"`
}

// A conjunction (logical AND) of predicates, which is used in evaluating a
// metrics filter. The operator must have at least two predicates in any combination,
// and an object must match all of the predicates for the filter to apply.
type AnalyticsAndOperator struct {
	Tags []*Tag `json:"tags,omitempty"`
}

// The filter used to describe a set of objects for analyses. A filter must
// have exactly one prefix, one tag, or one conjunction (AnalyticsAndOperator).
// If no filter is provided, all objects will be considered in any analysis.
type AnalyticsFilter struct {
	// A container of a key value name pair.
	Tag *Tag `json:"tag,omitempty"`
}

// Contains information about where to publish the analytics results.
type AnalyticsS3BucketDestination struct {
	Bucket          *string `json:"bucket,omitempty"`
	BucketAccountID *string `json:"bucketAccountID,omitempty"`
}

// Container for logging status information.
type BucketLoggingStatus struct {
	// Describes where logs are stored and the prefix that Amazon S3 assigns to
	// all log object keys for a bucket. For more information, see PUT Bucket logging
	// (https://docs.aws.amazon.com/AmazonS3/latest/API/RESTBucketPUTlogging.html)
	// in the Amazon Simple Storage Service API Reference.
	LoggingEnabled *LoggingEnabled `json:"loggingEnabled,omitempty"`
}

// In terms of implementation, a Bucket is a resource. An Amazon S3 bucket name
// is globally unique, and the namespace is shared by all AWS accounts.
type Bucket_SDK struct {
	CreationDate *metav1.Time `json:"creationDate,omitempty"`
	Name         *string      `json:"name,omitempty"`
}

// Describes the cross-origin access configuration for objects in an Amazon
// S3 bucket. For more information, see Enabling Cross-Origin Resource Sharing
// (https://docs.aws.amazon.com/AmazonS3/latest/dev/cors.html) in the Amazon
// Simple Storage Service Developer Guide.
type CORSConfiguration struct {
	CORSRules []*CORSRule `json:"corsRules,omitempty"`
}

// Specifies a cross-origin access rule for an Amazon S3 bucket.
type CORSRule struct {
	AllowedHeaders []*string `json:"allowedHeaders,omitempty"`
	AllowedMethods []*string `json:"allowedMethods,omitempty"`
	AllowedOrigins []*string `json:"allowedOrigins,omitempty"`
	ExposeHeaders  []*string `json:"exposeHeaders,omitempty"`
	MaxAgeSeconds  *int64    `json:"maxAgeSeconds,omitempty"`
}

// A container for describing a condition that must be met for the specified
// redirect to apply. For example, 1. If request is for pages in the /docs folder,
// redirect to the /documents folder. 2. If request results in HTTP error 4xx,
// redirect request to another host where you might process the error.
type Condition struct {
	HTTPErrorCodeReturnedEquals *string `json:"httpErrorCodeReturnedEquals,omitempty"`
	KeyPrefixEquals             *string `json:"keyPrefixEquals,omitempty"`
}

// The configuration information for the bucket.
type CreateBucketConfiguration struct {
	LocationConstraint *string `json:"locationConstraint,omitempty"`
}

// Information about the delete marker.
type DeleteMarkerEntry struct {
	Key *string `json:"key,omitempty"`
	// Container for the owner's display name and ID.
	Owner *Owner `json:"owner,omitempty"`
}

// Information about the deleted object.
type DeletedObject struct {
	Key *string `json:"key,omitempty"`
}

// Specifies information about where to publish analysis or configuration results
// for an Amazon S3 bucket and S3 Replication Time Control (S3 RTC).
type Destination struct {
	Account *string `json:"account,omitempty"`
	Bucket  *string `json:"bucket,omitempty"`
}

// Contains the type of server-side encryption used.
type Encryption struct {
	EncryptionType *string `json:"encryptionType,omitempty"`
	KMSKeyID       *string `json:"kmsKeyID,omitempty"`
}

// Container for all error elements.
type Error struct {
	Key *string `json:"key,omitempty"`
}

// The error information.
type ErrorDocument struct {
	Key *string `json:"key,omitempty"`
}

// Container for grant information.
type Grant struct {
	// Container for the person being granted permissions.
	Grantee *Grantee `json:"grantee,omitempty"`
}

// Container for the person being granted permissions.
type Grantee struct {
	DisplayName  *string `json:"displayName,omitempty"`
	EmailAddress *string `json:"emailAddress,omitempty"`
	ID           *string `json:"id,omitempty"`
	Type         *string `json:"type_,omitempty"`
	URI          *string `json:"uRI,omitempty"`
}

// Container for the Suffix element.
type IndexDocument struct {
	Suffix *string `json:"suffix,omitempty"`
}

// Container element that identifies who initiated the multipart upload.
type Initiator struct {
	DisplayName *string `json:"displayName,omitempty"`
	ID          *string `json:"id,omitempty"`
}

// A container for specifying S3 Intelligent-Tiering filters. The filters determine
// the subset of objects to which the rule applies.
type IntelligentTieringAndOperator struct {
	Tags []*Tag `json:"tags,omitempty"`
}

// The Filter is used to identify objects that the S3 Intelligent-Tiering configuration
// applies to.
type IntelligentTieringFilter struct {
	// A container of a key value name pair.
	Tag *Tag `json:"tag,omitempty"`
}

// Contains the bucket name, file format, bucket owner (optional), and prefix
// (optional) where inventory results are published.
type InventoryS3BucketDestination struct {
	AccountID *string `json:"accountID,omitempty"`
	Bucket    *string `json:"bucket,omitempty"`
}

// A lifecycle rule for individual objects in an Amazon S3 bucket.
type LifecycleRule struct {
	ID *string `json:"id,omitempty"`
}

// This is used in a Lifecycle Rule Filter to apply a logical AND to two or
// more predicates. The Lifecycle Rule will apply to any object matching all
// of the predicates configured inside the And operator.
type LifecycleRuleAndOperator struct {
	Tags []*Tag `json:"tags,omitempty"`
}

// The Filter is used to identify objects that a Lifecycle Rule applies to.
// A Filter must have exactly one of Prefix, Tag, or And specified.
type LifecycleRuleFilter struct {
	// A container of a key value name pair.
	Tag *Tag `json:"tag,omitempty"`
}

// Describes an Amazon S3 location that will receive the results of the restore
// request.
type Location struct {
	BucketName *string `json:"bucketName,omitempty"`
	// Container for TagSet elements.
	Tagging *Tagging `json:"tagging,omitempty"`
}

// Describes where logs are stored and the prefix that Amazon S3 assigns to
// all log object keys for a bucket. For more information, see PUT Bucket logging
// (https://docs.aws.amazon.com/AmazonS3/latest/API/RESTBucketPUTlogging.html)
// in the Amazon Simple Storage Service API Reference.
type LoggingEnabled struct {
	TargetBucket *string        `json:"targetBucket,omitempty"`
	TargetGrants []*TargetGrant `json:"targetGrants,omitempty"`
	TargetPrefix *string        `json:"targetPrefix,omitempty"`
}

// A conjunction (logical AND) of predicates, which is used in evaluating a
// metrics filter. The operator must have at least two predicates, and an object
// must match all of the predicates in order for the filter to apply.
type MetricsAndOperator struct {
	Tags []*Tag `json:"tags,omitempty"`
}

// Specifies a metrics configuration filter. The metrics configuration only
// includes objects that meet the filter's criteria. A filter must be a prefix,
// a tag, or a conjunction (MetricsAndOperator).
type MetricsFilter struct {
	// A container of a key value name pair.
	Tag *Tag `json:"tag,omitempty"`
}

// Container for the MultipartUpload for the Amazon S3 object.
type MultipartUpload struct {
	Key *string `json:"key,omitempty"`
	// Container for the owner's display name and ID.
	Owner *Owner `json:"owner,omitempty"`
}

// An object consists of data and its descriptive metadata.
type Object struct {
	Key *string `json:"key,omitempty"`
	// Container for the owner's display name and ID.
	Owner *Owner `json:"owner,omitempty"`
}

// Object Identifier is unique value to identify objects.
type ObjectIdentifier struct {
	Key *string `json:"key,omitempty"`
}

// The version of an object.
type ObjectVersion struct {
	Key *string `json:"key,omitempty"`
	// Container for the owner's display name and ID.
	Owner *Owner `json:"owner,omitempty"`
}

// Describes the location where the restore job's output is stored.
type OutputLocation struct {
	// Describes an Amazon S3 location that will receive the results of the restore
	// request.
	S3 *Location `json:"s3,omitempty"`
}

// Container for the owner's display name and ID.
type Owner struct {
	DisplayName *string `json:"displayName,omitempty"`
	ID          *string `json:"id,omitempty"`
}

// The container element for a bucket's ownership controls.
type OwnershipControls struct {
	Rules []*OwnershipControlsRule `json:"rules,omitempty"`
}

// The container element for an ownership control rule.
type OwnershipControlsRule struct {
	// The container element for object ownership for a bucket's ownership controls.
	//
	// BucketOwnerPreferred - Objects uploaded to the bucket change ownership to
	// the bucket owner if the objects are uploaded with the bucket-owner-full-control
	// canned ACL.
	//
	// ObjectWriter - The uploading account will own the object if the object is
	// uploaded with the bucket-owner-full-control canned ACL.
	ObjectOwnership *string `json:"objectOwnership,omitempty"`
}

// Specifies how requests are redirected. In the event of an error, you can
// specify a different error code to return.
type Redirect struct {
	HostName             *string `json:"hostName,omitempty"`
	HTTPRedirectCode     *string `json:"httpRedirectCode,omitempty"`
	Protocol             *string `json:"protocol,omitempty"`
	ReplaceKeyPrefixWith *string `json:"replaceKeyPrefixWith,omitempty"`
	ReplaceKeyWith       *string `json:"replaceKeyWith,omitempty"`
}

// Specifies the redirect behavior of all requests to a website endpoint of
// an Amazon S3 bucket.
type RedirectAllRequestsTo struct {
	HostName *string `json:"hostName,omitempty"`
	Protocol *string `json:"protocol,omitempty"`
}

// Specifies which Amazon S3 objects to replicate and where to store the replicas.
type ReplicationRule struct {
	ID *string `json:"id,omitempty"`
}

// A container for specifying rule filters. The filters determine the subset
// of objects to which the rule applies. This element is required only if you
// specify more than one filter.
//
// For example:
//
//    * If you specify both a Prefix and a Tag filter, wrap these filters in
//    an And tag.
//
//    * If you specify a filter based on multiple tags, wrap the Tag elements
//    in an And tag
type ReplicationRuleAndOperator struct {
	Tags []*Tag `json:"tags,omitempty"`
}

// A filter that identifies the subset of objects to which the replication rule
// applies. A Filter must specify exactly one Prefix, Tag, or an And child element.
type ReplicationRuleFilter struct {
	// A container of a key value name pair.
	Tag *Tag `json:"tag,omitempty"`
}

// Container for Payer.
type RequestPaymentConfiguration struct {
	Payer *string `json:"payer,omitempty"`
}

// Specifies the redirect behavior and when a redirect is applied. For more
// information about routing rules, see Configuring advanced conditional redirects
// (https://docs.aws.amazon.com/AmazonS3/latest/dev/how-to-page-redirect.html#advanced-conditional-redirects)
// in the Amazon Simple Storage Service Developer Guide.
type RoutingRule struct {
	// A container for describing a condition that must be met for the specified
	// redirect to apply. For example, 1. If request is for pages in the /docs folder,
	// redirect to the /documents folder. 2. If request results in HTTP error 4xx,
	// redirect request to another host where you might process the error.
	Condition *Condition `json:"condition,omitempty"`
	// Specifies how requests are redirected. In the event of an error, you can
	// specify a different error code to return.
	Redirect *Redirect `json:"redirect,omitempty"`
}

// Specifies lifecycle rules for an Amazon S3 bucket. For more information,
// see Put Bucket Lifecycle Configuration (https://docs.aws.amazon.com/AmazonS3/latest/API/RESTBucketPUTlifecycle.html)
// in the Amazon Simple Storage Service API Reference. For examples, see Put
// Bucket Lifecycle Configuration Examples (https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketLifecycleConfiguration.html#API_PutBucketLifecycleConfiguration_Examples)
type Rule struct {
	ID *string `json:"id,omitempty"`
}

// Specifies the use of SSE-KMS to encrypt delivered inventory reports.
type SSEKMS struct {
	KeyID *string `json:"keyID,omitempty"`
}

// Describes the default server-side encryption to apply to new objects in the
// bucket. If a PUT Object request doesn't specify any server-side encryption,
// this default encryption will be applied. For more information, see PUT Bucket
// encryption (https://docs.aws.amazon.com/AmazonS3/latest/API/RESTBucketPUTencryption.html)
// in the Amazon Simple Storage Service API Reference.
type ServerSideEncryptionByDefault struct {
	KMSMasterKeyID *string `json:"kmsMasterKeyID,omitempty"`
	SSEAlgorithm   *string `json:"sseAlgorithm,omitempty"`
}

// Specifies the default server-side-encryption configuration.
type ServerSideEncryptionConfiguration struct {
	Rules []*ServerSideEncryptionRule `json:"rules,omitempty"`
}

// Specifies the default server-side encryption configuration.
type ServerSideEncryptionRule struct {
	// Describes the default server-side encryption to apply to new objects in the
	// bucket. If a PUT Object request doesn't specify any server-side encryption,
	// this default encryption will be applied. For more information, see PUT Bucket
	// encryption (https://docs.aws.amazon.com/AmazonS3/latest/API/RESTBucketPUTencryption.html)
	// in the Amazon Simple Storage Service API Reference.
	ApplyServerSideEncryptionByDefault *ServerSideEncryptionByDefault `json:"applyServerSideEncryptionByDefault,omitempty"`
	BucketKeyEnabled                   *bool                          `json:"bucketKeyEnabled,omitempty"`
}

// A container of a key value name pair.
type Tag struct {
	Key   *string `json:"key,omitempty"`
	Value *string `json:"value,omitempty"`
}

// Container for TagSet elements.
type Tagging struct {
	TagSet []*Tag `json:"tagSet,omitempty"`
}

// Container for granting information.
type TargetGrant struct {
	// Container for the person being granted permissions.
	Grantee    *Grantee `json:"grantee,omitempty"`
	Permission *string  `json:"permission,omitempty"`
}

// Specifies website configuration parameters for an Amazon S3 bucket.
type WebsiteConfiguration struct {
	// The error information.
	ErrorDocument *ErrorDocument `json:"errorDocument,omitempty"`
	// Container for the Suffix element.
	IndexDocument *IndexDocument `json:"indexDocument,omitempty"`
	// Specifies the redirect behavior of all requests to a website endpoint of
	// an Amazon S3 bucket.
	RedirectAllRequestsTo *RedirectAllRequestsTo `json:"redirectAllRequestsTo,omitempty"`
	RoutingRules          []*RoutingRule         `json:"routingRules,omitempty"`
}
