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

package bucket_test

import (
	"testing"

	bucket "github.com/aws-controllers-k8s/s3-controller/pkg/resource/bucket"
	"github.com/stretchr/testify/assert"
)

func Test_IsDirectoryBucketName(t *testing.T) {
	assert := assert.New(t)

	assert.True(bucket.IsDirectoryBucketName("my-bucket--usw2-az5--x-s3"))
	assert.True(bucket.IsDirectoryBucketName("my-bucket--use1-az2--x-s3"))

	assert.False(bucket.IsDirectoryBucketName("my-regular-bucket"))
	assert.False(bucket.IsDirectoryBucketName("bucket-with--use1-az2-zone"))
}
