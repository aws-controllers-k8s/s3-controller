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
	"testing"

	svcsdk "github.com/aws/aws-sdk-go-v2/service/s3"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/s3/types"

	svcapitypes "github.com/aws-controllers-k8s/s3-controller/apis/v1alpha1"
)

func TestNotificationEventBridgeConfigurationRoundTrip(t *testing.T) {
	rm := &resourceManager{}

	enabled := &resource{ko: &svcapitypes.Bucket{}}
	enabled.ko.Spec.Notification = &svcapitypes.NotificationConfiguration{
		EventBridgeConfiguration: map[string]*string{},
	}
	if rm.newNotificationConfiguration(enabled).EventBridgeConfiguration == nil {
		t.Error("an eventBridgeConfiguration in the spec must be sent to PutBucketNotificationConfiguration")
	}

	disabled := &resource{ko: &svcapitypes.Bucket{}}
	disabled.ko.Spec.Notification = &svcapitypes.NotificationConfiguration{}
	if rm.newNotificationConfiguration(disabled).EventBridgeConfiguration != nil {
		t.Error("a spec without eventBridgeConfiguration must not enable it")
	}

	observed := rm.setResourceNotification(enabled, &svcsdk.GetBucketNotificationConfigurationOutput{
		EventBridgeConfiguration: &svcsdktypes.EventBridgeConfiguration{},
	})
	if observed.EventBridgeConfiguration == nil {
		t.Error("an EventBridge configuration reported by AWS must be read back into the spec")
	}
	if rm.setResourceNotification(enabled, &svcsdk.GetBucketNotificationConfigurationOutput{}).EventBridgeConfiguration != nil {
		t.Error("a bucket without EventBridge notifications must read back without the field")
	}

	desired := enabled
	latest := &resource{ko: &svcapitypes.Bucket{}}
	latest.ko.Spec.Notification = observed
	if newResourceDelta(desired, latest).DifferentAt("Spec.Notification.EventBridgeConfiguration") {
		t.Error("an enabled configuration must not differ from the read back")
	}
	if !newResourceDelta(desired, disabled).DifferentAt("Spec.Notification.EventBridgeConfiguration") {
		t.Error("enabling EventBridge must register as a difference")
	}
}
