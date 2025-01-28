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
	"fmt"
	"testing"

	bucket "github.com/aws-controllers-k8s/s3-controller/pkg/resource/bucket"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/s3"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
)

var (
	OwnerDisplayName = "my-test-user"
	OwnerID          = "123456789"
	RandomGranteeURI = "http://my-random-grantee.example.com/lol"
)

func s(s string) *string { return &s }

func provideOwner() *svcsdktypes.Owner {
	return &svcsdktypes.Owner{
		DisplayName: &OwnerDisplayName,
		ID:          &OwnerID,
	}
}

func provideOwnerGrantee() *svcsdktypes.Grantee {
	return &svcsdktypes.Grantee{
		DisplayName: &OwnerDisplayName,
		ID:          &OwnerID,
		Type:        svcsdktypes.TypeCanonicalUser,
	}
}

func provideMockUserFullControl() []svcsdktypes.Grant {
	return []svcsdktypes.Grant{
		{
			Grantee:    provideOwnerGrantee(),
			Permission: svcsdktypes.PermissionFullControl,
		},
	}
}

func wrapGrants(grants []svcsdktypes.Grant) *svcsdk.GetBucketAclOutput {
	return &svcsdk.GetBucketAclOutput{
		Grants: grants,
		Owner:  provideOwner(),
	}
}

func cannedPrivateOutput() *svcsdk.GetBucketAclOutput {
	return wrapGrants(provideMockUserFullControl())
}

func cannedLogDeliveryOutput() *svcsdk.GetBucketAclOutput {
	grants := provideMockUserFullControl()
	logDeliveryGrantee := &svcsdktypes.Grantee{
		Type: svcsdktypes.TypeGroup,
		URI:  &bucket.GranteeLogDeliveryURI,
	}
	writeGrant := svcsdktypes.Grant{
		Grantee:    logDeliveryGrantee,
		Permission: svcsdktypes.PermissionWrite,
	}
	readACPGrant := svcsdktypes.Grant{
		Grantee:    logDeliveryGrantee,
		Permission: svcsdktypes.PermissionReadAcp,
	}
	grants = append(grants, writeGrant)
	grants = append(grants, readACPGrant)

	return wrapGrants(grants)
}

func cannedPublicReadWriteOutput() *svcsdk.GetBucketAclOutput {
	grants := provideMockUserFullControl()
	allUsersGrantee := &svcsdktypes.Grantee{
		Type: svcsdktypes.TypeGroup,
		URI:  &bucket.GranteeAllUsersURI,
	}
	writeGrant := svcsdktypes.Grant{
		Grantee:    allUsersGrantee,
		Permission: svcsdktypes.PermissionWrite,
	}
	readGrant := svcsdktypes.Grant{
		Grantee:    allUsersGrantee,
		Permission: svcsdktypes.PermissionRead,
	}
	grants = append(grants, writeGrant)
	grants = append(grants, readGrant)

	return wrapGrants(grants)
}

func allGrantsOutput() *svcsdk.GetBucketAclOutput {
	grants := provideMockUserFullControl()
	randomGrantee := &svcsdktypes.Grantee{
		Type: svcsdktypes.TypeGroup,
		URI:  &RandomGranteeURI,
	}
	writeGrant := svcsdktypes.Grant{
		Grantee:    randomGrantee,
		Permission: svcsdktypes.PermissionWrite,
	}
	writeACPGrant := svcsdktypes.Grant{
		Grantee:    randomGrantee,
		Permission: svcsdktypes.PermissionWriteAcp,
	}
	readGrant := svcsdktypes.Grant{
		Grantee:    randomGrantee,
		Permission: svcsdktypes.PermissionRead,
	}
	readACPGrant := svcsdktypes.Grant{
		Grantee:    randomGrantee,
		Permission: svcsdktypes.PermissionReadAcp,
	}
	grants = append(grants, writeGrant)
	grants = append(grants, writeACPGrant)
	grants = append(grants, readGrant)
	grants = append(grants, readACPGrant)

	return wrapGrants(grants)
}

func multiplePermissionGrantsOutput() *svcsdk.GetBucketAclOutput {
	grants := provideMockUserFullControl()
	anotherFulLControl := svcsdktypes.Grant{
		Grantee: &svcsdktypes.Grantee{
			Type: svcsdktypes.TypeGroup,
			URI:  s(RandomGranteeURI),
		},
		Permission: svcsdktypes.PermissionFullControl,
	}

	grants = append(grants, anotherFulLControl)
	return wrapGrants(grants)
}

func Test_GetHeadersFromGrants(t *testing.T) {
	assert := assert.New(t)

	privateGrants := cannedPrivateOutput()
	headers := bucket.GetHeadersFromGrants(privateGrants)
	assert.Equal(headers.FullControl, fmt.Sprintf("id=%s", OwnerID))
	assert.Empty(headers.Read)
	assert.Empty(headers.ReadACP)
	assert.Empty(headers.Write)
	assert.Empty(headers.WriteACP)

	logDeliveryGrants := cannedLogDeliveryOutput()
	headers = bucket.GetHeadersFromGrants(logDeliveryGrants)
	assert.Equal(headers.FullControl, fmt.Sprintf("id=%s", OwnerID))
	assert.Empty(headers.Read)
	assert.Equal(headers.ReadACP, fmt.Sprintf("uri=%s", bucket.GranteeLogDeliveryURI))
	assert.Equal(headers.Write, fmt.Sprintf("uri=%s", bucket.GranteeLogDeliveryURI))
	assert.Empty(headers.WriteACP)

	allGrants := allGrantsOutput()
	headers = bucket.GetHeadersFromGrants(allGrants)
	assert.Equal(headers.FullControl, fmt.Sprintf("id=%s", OwnerID))
	assert.Equal(headers.Read, fmt.Sprintf("uri=%s", RandomGranteeURI))
	assert.Equal(headers.ReadACP, fmt.Sprintf("uri=%s", RandomGranteeURI))
	assert.Equal(headers.Write, fmt.Sprintf("uri=%s", RandomGranteeURI))
	assert.Equal(headers.WriteACP, fmt.Sprintf("uri=%s", RandomGranteeURI))

	multiplePermissionGrants := multiplePermissionGrantsOutput()
	headers = bucket.GetHeadersFromGrants(multiplePermissionGrants)
	assert.Equal(headers.FullControl, fmt.Sprintf("id=%s,uri=%s", OwnerID, RandomGranteeURI))
	assert.Empty(headers.Read)
	assert.Empty(headers.ReadACP)
	assert.Empty(headers.Write)
	assert.Empty(headers.WriteACP)
}

func Test_GetPossibleCannedACLsFromGrants(t *testing.T) {
	assert := assert.New(t)

	privateGrants := cannedPrivateOutput()
	possibilities := bucket.GetPossibleCannedACLsFromGrants(privateGrants)
	assert.ElementsMatch(possibilities, []string{bucket.CannedACLPrivate, bucket.CannedBucketOwnerRead, bucket.CannedBucketOwnerFullControl})

	logDeliveryGrants := cannedLogDeliveryOutput()
	possibilities = bucket.GetPossibleCannedACLsFromGrants(logDeliveryGrants)
	assert.ElementsMatch(possibilities, []string{bucket.CannedLogDeliveryWrite})

	publicReadWriteGrants := cannedPublicReadWriteOutput()
	possibilities = bucket.GetPossibleCannedACLsFromGrants(publicReadWriteGrants)
	assert.ElementsMatch(possibilities, []string{bucket.CannedPublicReadWrite})
}
