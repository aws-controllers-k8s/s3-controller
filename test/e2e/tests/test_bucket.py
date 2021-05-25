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

"""Integration tests for the S3 Bucket API.
"""

import boto3
import pytest
import time
import logging
from typing import Dict, Tuple

from acktest.resources import random_suffix_name
from acktest.k8s import resource as k8s
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_s3_resource
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.bootstrap_resources import TestBootstrapResources, get_bootstrap_resources

RESOURCE_PLURAL = "buckets"

CREATE_WAIT_AFTER_SECONDS = 10
DELETE_WAIT_AFTER_SECONDS = 10

@pytest.fixture(scope="module")
def s3_client():
    return boto3.client("s3")

@service_marker
@pytest.mark.canary
class TestBucket:
    def get_bucket(self, s3_client, bucket_name: str) -> dict:
        try:
            resp = s3_client.list_buckets()
        except Exception as e:
            logging.debug(e)
            return None

        buckets = resp["Buckets"]
        for bucket in buckets:
            if bucket["Name"] == bucket_name:
                return bucket

        return None

    def bucket_exists(self, s3_client, bucket_name: str) -> bool:
        return self.get_bucket(s3_client, bucket_name) is not None

    def test_smoke(self, s3_client):
        resource_name = random_suffix_name("s3-bucket", 24)

        replacements = REPLACEMENT_VALUES.copy()
        replacements["BUCKET_NAME"] = resource_name

        # Load Bucket CR
        resource_data = load_s3_resource(
            "bucket",
            additional_replacements=replacements,
        )
        logging.debug(resource_data)

        # Create k8s resource
        ref = k8s.CustomResourceReference(
            CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL,
            resource_name, namespace="default",
        )
        k8s.create_custom_resource(ref, resource_data)
        cr = k8s.wait_resource_consumed_by_controller(ref)

        assert cr is not None
        assert k8s.get_resource_exists(ref)

        time.sleep(CREATE_WAIT_AFTER_SECONDS)

        # Check S3 Bucket exists
        exists = self.bucket_exists(s3_client, resource_name)
        assert exists

        # Delete k8s resource
        _, deleted = k8s.delete_custom_resource(ref)
        assert deleted is True

        time.sleep(DELETE_WAIT_AFTER_SECONDS)

        # Check S3 Bucket doesn't exists
        exists = self.bucket_exists(s3_client, resource_name)
        assert not exists