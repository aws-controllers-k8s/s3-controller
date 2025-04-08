# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
# 	 http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.

"""Integration tests for the deletion policy annotation on Bucket.
"""

from enum import Enum
import pytest
import time
import logging

from acktest.resources import random_suffix_name
from acktest.k8s import resource as k8s
from acktest import tags as tags
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_s3_resource
from e2e.tests.test_bucket import bucket_exists, get_bucket
from e2e.replacement_values import REPLACEMENT_VALUES

CREATE_WAIT_AFTER_SECONDS = 10
MODIFY_WAIT_AFTER_SECONDS = 10
DELETE_WAIT_AFTER_SECONDS = 10
ACK_SYSTEM_TAG_PREFIX = "services.k8s.aws/"
AWS_SYSTEM_TAG_PREFIX = "aws:"

class AdoptionPolicy(str, Enum):
    NONE = ""
    ADOPT = "adopt"
    ADOPT_OR_CREATE = "adopt-or-create"


@pytest.fixture
def bucket_adoption_policy(request, s3_client):
    replacements = REPLACEMENT_VALUES.copy()
    bucket_name = replacements["ADOPTION_BUCKET_NAME"]
    
    replacements["ADOPTION_POLICY"] = AdoptionPolicy.ADOPT
    replacements["ADOPTION_FIELDS"] = f'{{\\\"name\\\": \\\"{bucket_name}\\\"}}'
    replacements["BUCKET_NAME"] = bucket_name

    filename = ""

    resource_name = ""

    marker = request.node.get_closest_marker("resource_data")
    assert marker is not None
    data = marker.args[0]
    assert 'adoption-policy' in data
    replacements["ADOPTION_POLICY"] = data['adoption-policy']
    assert 'filename' in data
    filename = data['filename']
    assert 'resource-name' in data
    resource_name = random_suffix_name(data['resource-name'], 32)
    replacements["RANDOM_BUCKET_NAME"] = resource_name

    resource_data = load_s3_resource(
        filename,
        additional_replacements=replacements,
    )

    # Create k8s resource
    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, "buckets",
        resource_name, namespace="default")
    k8s.create_custom_resource(ref, resource_data)

    time.sleep(CREATE_WAIT_AFTER_SECONDS)
    cr = k8s.wait_resource_consumed_by_controller(ref)

    k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=5)
    cr = k8s.get_resource(ref)
    assert cr is not None

    yield (ref, cr)

    _, deleted = k8s.delete_custom_resource(ref, DELETE_WAIT_AFTER_SECONDS)
    assert deleted

@pytest.fixture(scope="module")
def adopt_stack_bucket(s3_client):
    replacements = REPLACEMENT_VALUES.copy()
    bucket_name = replacements["STACK_BUCKET_NAME"]
    replacements["ADOPTION_POLICY"] = AdoptionPolicy.ADOPT
    replacements["ADOPTION_FIELDS"] = f'{{\\\"name\\\": \\\"{bucket_name}\\\"}}'

    resource_data = load_s3_resource(
        "bucket_adoption_stack",
        additional_replacements=replacements,
    )

    # Create k8s resource
    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, "buckets",
        bucket_name, namespace="default")
    k8s.create_custom_resource(ref, resource_data)

    time.sleep(CREATE_WAIT_AFTER_SECONDS)
    cr = k8s.wait_resource_consumed_by_controller(ref)

    assert cr is not None
    assert k8s.get_resource_exists(ref)

    yield (ref, cr)

    _, deleted = k8s.delete_custom_resource(ref, DELETE_WAIT_AFTER_SECONDS)
    assert deleted


@service_marker
@pytest.mark.canary
class TestAdoptionPolicyBucket:
    @pytest.mark.resource_data({'adoption-policy': AdoptionPolicy.ADOPT, 'filename': 'bucket_adopt', 'resource-name': 'adopt'})
    def test_adopt_policy(
        self, s3_client, bucket_adoption_policy, s3_resource
    ):
        (ref, cr) = bucket_adoption_policy

        # Spec will be added by controller
        assert 'spec' in cr
        assert 'name' in cr['spec']
        bucket_name = cr['spec']['name']

        updates = {
            "spec": {
                "versioning": {
                    "status": "Suspended"
                },
            }
        }

        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        cr = k8s.wait_resource_consumed_by_controller(ref)
        assert cr is not None
        assert 'spec' in cr
        assert 'versioning' in cr['spec']
        assert 'status' in cr['spec']['versioning']
        status = cr['spec']['versioning']['status']
        latest = get_bucket(s3_resource, bucket_name)
        assert latest is not None
        versioning = latest.Versioning()
        assert versioning.status == status
    
    @pytest.mark.resource_data({'adoption-policy': AdoptionPolicy.ADOPT_OR_CREATE, 'filename': 'bucket_adopt_or_create', 'resource-name': 'adopt-or-create'})
    def test_adopt_or_create_policy(
        self, s3_client, bucket_adoption_policy, s3_resource
    ):
        (ref, cr) = bucket_adoption_policy

        # Spec will be added by controller
        k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=5)
        assert 'spec' in cr
        assert 'name' in cr['spec']
        bucket_name = cr['spec']['name']

        latest = get_bucket(s3_resource, bucket_name)
        assert latest is not None
        tagging = latest.Tagging()

        initial_tags = {
            "tag_key": "tag_value"
        }
        tags.assert_ack_system_tags(
            tags=tagging.tag_set,
        )
        tags.assert_equal_without_ack_tags(
            expected=initial_tags,
            actual=tagging.tag_set,
        )


    @pytest.mark.resource_data({'adoption-policy': AdoptionPolicy.ADOPT_OR_CREATE, 'filename': 'bucket_adopt_or_create_not_exist', 'resource-name': 'adopt-or-create-not-exist'})
    def test_adopt_or_create_policy_non_existent(
        self, s3_client, bucket_adoption_policy, s3_resource
    ):
        (ref, cr) = bucket_adoption_policy

        # Spec will be added by controller
        assert 'spec' in cr
        assert 'name' in cr['spec']
        k8s.wait_on_condition(ref, "ACK.ResourceSynced", "True", wait_periods=5)

        name = cr['spec']['name']
        latest = get_bucket(s3_resource, name)
        assert latest is not None

    def test_adoption_update_tags(
        self, s3_client, adopt_stack_bucket, s3_resource
    ):
        (ref, cr) = adopt_stack_bucket

        # Spec will be added by controller
        assert 'spec' in cr
        assert 'name' in cr['spec']
        assert 'tagSet' not in cr['spec']['tagging']
        bucket_name = cr['spec']['name']

        updates = {
            "spec": {
                "tagging": {
                    "tagSet": [
                        {"key": "newKey", "value": "newVal"}
                    ]
                }
            }
        }

        k8s.patch_custom_resource(ref, updates)
        time.sleep(MODIFY_WAIT_AFTER_SECONDS)

        cr = k8s.wait_resource_consumed_by_controller(ref)
        assert cr is not None
        assert 'spec' in cr
        assert 'tagging' in cr['spec']
        assert 'tagSet' in cr['spec']['tagging']

        latest = get_bucket(s3_resource, bucket_name)
        assert latest is not None
        tagging = latest.Tagging()

        latest = cleanTags(tagging.tag_set)
        # +2 here because we want to see if we're also filtering
        # through the aws tags, besides just the ack tags
        assert len(tagging.tag_set) > len(latest) + 2
        desired = cr['spec']['tagging']['tagSet']
        for i in range(1):
            assert desired[i]["key"] == latest[i]["Key"]
            assert desired[i]["value"] == latest[i]["Value"]


def cleanTags(tags: list,
          key_member_name: str = 'Key',
    ) -> list:
    if isinstance(tags, list):
        return [
            t for t in tags if not t[key_member_name].startswith(AWS_SYSTEM_TAG_PREFIX) 
                and not t[key_member_name].startswith(ACK_SYSTEM_TAG_PREFIX)
        ]
    else:
        raise RuntimeError('tags parameter can only be list type')