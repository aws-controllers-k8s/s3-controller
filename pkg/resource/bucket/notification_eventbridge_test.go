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
	"testing"

	svcsdk "github.com/aws/aws-sdk-go-v2/service/s3"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/require"

	svcapitypes "github.com/aws-controllers-k8s/s3-controller/apis/v1alpha1"
)

func bucketWithNotification(n *svcapitypes.NotificationConfiguration) *resource {
	r := newBucketResource("my-eventbridge-bucket")
	r.ko.Spec.Notification = n
	return r
}

func Test_newNotificationConfiguration_EventBridgeEnabled(t *testing.T) {
	rm := &resourceManager{}
	r := bucketWithNotification(&svcapitypes.NotificationConfiguration{EventBridgeConfiguration: map[string]*string{}})
	require.NotNil(t, rm.newNotificationConfiguration(r).EventBridgeConfiguration)
}

func Test_newNotificationConfiguration_EventBridgeUnset(t *testing.T) {
	rm := &resourceManager{}
	r := bucketWithNotification(&svcapitypes.NotificationConfiguration{})
	require.Nil(t, rm.newNotificationConfiguration(r).EventBridgeConfiguration)
}

func Test_setResourceNotification_EventBridgeReadBack(t *testing.T) {
	rm := &resourceManager{}
	r := bucketWithNotification(&svcapitypes.NotificationConfiguration{EventBridgeConfiguration: map[string]*string{}})
	observed := rm.setResourceNotification(r, &svcsdk.GetBucketNotificationConfigurationOutput{
		EventBridgeConfiguration: &svcsdktypes.EventBridgeConfiguration{},
	})
	require.NotNil(t, observed.EventBridgeConfiguration)
	require.Nil(t, rm.setResourceNotification(r, &svcsdk.GetBucketNotificationConfigurationOutput{}).EventBridgeConfiguration)
}

func Test_newResourceDelta_EventBridgeConfiguration(t *testing.T) {
	enabled := bucketWithNotification(&svcapitypes.NotificationConfiguration{EventBridgeConfiguration: map[string]*string{}})
	disabled := bucketWithNotification(&svcapitypes.NotificationConfiguration{})
	require.False(t, newResourceDelta(enabled, enabled).DifferentAt("Spec.Notification.EventBridgeConfiguration"))
	require.True(t, newResourceDelta(enabled, disabled).DifferentAt("Spec.Notification.EventBridgeConfiguration"))
}

// An EventBridge-only configuration has to be read back through
// addPutFieldsToSpec as well, or every reconcile sees a difference and calls
// PutBucketNotificationConfiguration again.
func Test_addPutFieldsToSpec_NotificationEventBridgeOnly(t *testing.T) {
	require := require.New(t)

	rm := &resourceManager{
		sdkapi: newMockedSDKClient(map[string]opResult{
			"GetBucketNotificationConfiguration": {output: &svcsdk.GetBucketNotificationConfigurationOutput{
				EventBridgeConfiguration: &svcsdktypes.EventBridgeConfiguration{},
			}},
		}),
	}

	desired := bucketWithNotification(&svcapitypes.NotificationConfiguration{EventBridgeConfiguration: map[string]*string{}})
	ko := desired.ko.DeepCopy()

	require.NoError(rm.addPutFieldsToSpec(context.Background(), desired, ko))
	require.NotNil(ko.Spec.Notification)
	require.NotNil(ko.Spec.Notification.EventBridgeConfiguration)
	require.False(newResourceDelta(desired, &resource{ko: ko}).DifferentAt("Spec.Notification"))
}
