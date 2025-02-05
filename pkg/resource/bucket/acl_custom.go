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
	"fmt"
	"strings"

	svcsdk "github.com/aws/aws-sdk-go-v2/service/s3"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Only some of these exist in the SDK, so duplicating them all here
var (
	CannedACLPrivate             = "private"
	CannedPublicRead             = "public-read"
	CannedPublicReadWrite        = "public-read-write"
	CannedAWSExecRead            = "aws-exec-read"
	CannedAuthenticatedRead      = "authenticated-read"
	CannedBucketOwnerRead        = "bucket-owner-read"
	CannedBucketOwnerFullControl = "bucket-owner-full-control"
	CannedLogDeliveryWrite       = "log-delivery-write"
)

var (
	GranteeZATeamID              = "6aa5a366c34c1cbe25dc49211496e913e0351eb0e8c37aa3477e40942ec6b97c"
	GranteeLogDeliveryURI        = "http://acs.amazonaws.com/groups/s3/LogDelivery"
	GranteeAllUsersURI           = "http://acs.amazonaws.com/groups/global/AllUsers"
	GranteeAuthenticatedUsersURI = "http://acs.amazonaws.com/groups/global/AuthenticatedUsers"
)

var (
	HeaderUserIDFormat = "id=%s"
	HeaderURIFormat    = "uri=%s"
)

type aclGrantHeaders struct {
	FullControl string
	Read        string
	ReadACP     string
	Write       string
	WriteACP    string
}

// hasOwnerFullControl returns true if any of the grants matches the owner
// and has full control permissions.
func hasOwnerFullControl(owner *svcsdktypes.Owner, grants []svcsdktypes.Grant) bool {
	for _, grant := range grants {
		if grant.Grantee == nil ||
			grant.Grantee.ID == nil ||
			*grant.Grantee.ID != *owner.ID {
			continue
		}

		return grant.Permission == svcsdktypes.PermissionFullControl
	}
	return false
}

// grantsContainPermission will return true if any of the grants have the
// permission matching the one supplied.
func grantsContainPermission(permission svcsdktypes.Permission, grants []*svcsdktypes.Grant) bool {
	for _, grant := range grants {
		if grant.Permission == svcsdktypes.Permission(permission) {
			return true
		}
	}
	return false
}

// getGrantsByGroupURI searches a list of ACL grants for any that have a
// group type grantee with the given URI.
func getGrantsByGroupURI(uri string, grants []svcsdktypes.Grant) []*svcsdktypes.Grant {
	matching := []*svcsdktypes.Grant{}

	for _, grant := range grants {
		if grant.Grantee == nil {
			continue
		}

		if grant.Grantee.Type != svcsdktypes.TypeGroup {
			continue
		}

		if *grant.Grantee.URI == uri {
			matching = append(matching, &grant)
		}
	}
	return matching
}

// getGrantsByCanonicalUserID searches a list of ACL grants for any that have a
// canonical user type grantee with the given ID.
func getGrantsByCanonicalUserID(id string, grants []svcsdktypes.Grant) []*svcsdktypes.Grant {
	matching := []*svcsdktypes.Grant{}

	for _, grant := range grants {
		if grant.Grantee == nil {
			continue
		}

		if grant.Grantee.Type != svcsdktypes.TypeCanonicalUser {
			continue
		}

		if *grant.Grantee.ID == id {
			matching = append(matching, &grant)
		}
	}
	return matching
}

// getGrantsByPermission searches a list of ACL grants for any that have the
// given permission.
func getGrantsByPermission(permission string, grants []svcsdktypes.Grant) []*svcsdktypes.Grant {
	matching := []*svcsdktypes.Grant{}

	for _, grant := range grants {
		if grant.Permission == svcsdktypes.Permission(permission) {
			matching = append(matching, &grant)
		}
	}
	return matching
}

// formGrantHeader will form a grant header string from a list of grants
func formGrantHeader(grants []*svcsdktypes.Grant) string {
	headers := []string{}
	for _, grant := range grants {
		if grant.Grantee == nil {
			continue
		}

		if grant.Grantee.Type == svcsdktypes.TypeGroup {
			headers = append(headers, fmt.Sprintf(HeaderURIFormat, *grant.Grantee.URI))
		}
		if grant.Grantee.Type == svcsdktypes.TypeCanonicalUser {
			headers = append(headers, fmt.Sprintf(HeaderUserIDFormat, *grant.Grantee.ID))
		}
	}
	return strings.Join(headers, ",")
}

// isDefaultCannedACLPossibilities determines whether the list of joined ACL
// possibilites is the default for a bucket.
func isDefaultCannedACLPossibilities(joinedPossibilities string) bool {
	return matchPossibleCannedACL(CannedACLPrivate, joinedPossibilities) != nil
}

// matchPossibleCannedACL attempts to find a canned ACL string in a joined
// list of possibilities. If any of the possibilities matches, it will be
// returned, otherwise nil.
func matchPossibleCannedACL(search string, joinedPossibilities string) *string {
	splitPossibilities := strings.Split(joinedPossibilities, CannedACLJoinDelimiter)
	for _, possible := range splitPossibilities {
		if search == possible {
			return &possible
		}
	}
	return nil
}

// GetHeadersFromGrants will return a list of grant headers from grants
func GetHeadersFromGrants(
	resp *svcsdk.GetBucketAclOutput,
) aclGrantHeaders {
	headers := aclGrantHeaders{
		FullControl: formGrantHeader(getGrantsByPermission(string(svcsdktypes.PermissionFullControl), resp.Grants)),
		Read:        formGrantHeader(getGrantsByPermission(string(svcsdktypes.PermissionRead), resp.Grants)),
		ReadACP:     formGrantHeader(getGrantsByPermission(string(svcsdktypes.PermissionReadAcp), resp.Grants)),
		Write:       formGrantHeader(getGrantsByPermission(string(svcsdktypes.PermissionWrite), resp.Grants)),
		WriteACP:    formGrantHeader(getGrantsByPermission(string(svcsdktypes.PermissionWriteAcp), resp.Grants)),
	}

	return headers
}

// GetPossibleCannedACLsFromGrants will return a list of canned ACLs that match
// the list of grants. This method will return nil if the grants did not match
// any canned ACLs.
func GetPossibleCannedACLsFromGrants(
	resp *svcsdk.GetBucketAclOutput,
) []string {
	owner := resp.Owner
	grants := resp.Grants

	// All canned ACLs include a grant with owner full control
	if !hasOwnerFullControl(owner, grants) {
		return []string{}
	}

	switch len(grants) {
	case 1:
		return []string{CannedACLPrivate, CannedBucketOwnerRead, CannedBucketOwnerFullControl}
	case 2:
		execTeamGrant := getGrantsByCanonicalUserID(GranteeZATeamID, grants)
		if grantsContainPermission(svcsdktypes.PermissionRead, execTeamGrant) {
			return []string{CannedAWSExecRead}
		}

		allUsersGrants := getGrantsByGroupURI(GranteeAllUsersURI, grants)
		if grantsContainPermission(svcsdktypes.PermissionRead, allUsersGrants) {
			return []string{CannedPublicRead}
		}

		authenticatedUsersGrants := getGrantsByGroupURI(GranteeAuthenticatedUsersURI, grants)
		if grantsContainPermission(svcsdktypes.PermissionRead, authenticatedUsersGrants) {
			return []string{CannedAuthenticatedRead}
		}
	case 3:
		logDeliveryGrants := getGrantsByGroupURI(GranteeLogDeliveryURI, grants)
		if grantsContainPermission(svcsdktypes.PermissionWrite, logDeliveryGrants) &&
			grantsContainPermission(svcsdktypes.PermissionReadAcp, logDeliveryGrants) {
			return []string{CannedLogDeliveryWrite}
		}

		allUsersGrants := getGrantsByGroupURI(GranteeAllUsersURI, grants)
		if grantsContainPermission(svcsdktypes.PermissionRead, allUsersGrants) &&
			grantsContainPermission(svcsdktypes.PermissionWrite, allUsersGrants) {
			return []string{CannedPublicReadWrite}
		}
	}

	return []string{}
}
