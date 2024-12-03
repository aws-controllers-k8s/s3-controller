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
from acktest import adoption as adoption
from acktest import tags as tags
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_s3_resource
from e2e.tests.test_bucket import bucket_exists, get_bucket
from e2e.replacement_values import REPLACEMENT_VALUES

CREATE_WAIT_AFTER_SECONDS = 10
MODIFY_WAIT_AFTER_SECONDS = 10
DELETE_WAIT_AFTER_SECONDS = 10

class AdoptionPolicy(str, Enum):
    NONE = ""
    ADOPT = "adopt"
    ADOPT_OR_CREATE = "adopt-or-create"


@pytest.fixture(scope="module")
def adoption_policy_adopt_bucket(s3_client):
    replacements = REPLACEMENT_VALUES.copy()
    bucket_name = replacements["ADOPTION_BUCKET_NAME"]
    replacements["ADOPTION_POLICY"] = AdoptionPolicy.ADOPT
    replacements["ADOPTION_FIELDS"] = f'{{\\\"name\\\": \\\"{bucket_name}\\\"}}'

    resource_data = load_s3_resource(
        "bucket_adoption_policy",
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
    def test_adoption_policy(
        self, s3_client, adoption_policy_adopt_bucket, s3_resource
    ):
        (ref, cr) = adoption_policy_adopt_bucket

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


