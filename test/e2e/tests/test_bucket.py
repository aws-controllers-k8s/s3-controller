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

import pytest
import time
import logging
import re
import boto3
from typing import  Generator
from dataclasses import dataclass

from acktest.resources import random_suffix_name
from acktest.k8s import resource as k8s
from acktest.aws.identity import get_region
from acktest import adoption as adoption
from acktest import tags as tags
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_s3_resource
from e2e.replacement_values import REPLACEMENT_VALUES

RESOURCE_KIND = "Bucket"
RESOURCE_PLURAL = "buckets"

CREATE_WAIT_AFTER_SECONDS = 10
MODIFY_WAIT_AFTER_SECONDS = 10
DELETE_WAIT_AFTER_SECONDS = 10

@dataclass
class Bucket:
    ref: k8s.CustomResourceReference
    resource_name: str
    resource_data: str

def get_bucket(s3_resource, bucket_name: str):
    return s3_resource.Bucket(bucket_name)

def bucket_exists(s3_client, bucket: Bucket) -> bool:
    try:
        resp = s3_client.list_buckets()
    except Exception as e:
        logging.debug(e)
        return False

    buckets = resp["Buckets"]
    for _bucket in buckets:
        if _bucket["Name"] == bucket.resource_name:
            return True

    return False

def load_bucket_resource(resource_file_name: str, resource_name: str, additional_replacements: dict = {}):
    replacements = {**REPLACEMENT_VALUES.copy(), **additional_replacements}
    replacements["BUCKET_NAME"] = resource_name

    resource_data = load_s3_resource(
        resource_file_name,
        additional_replacements=replacements,
    )
    logging.debug(resource_data)
    return resource_data

def create_bucket(resource_file_name: str, namespace: str = "default", additional_replacements: dict = {}) -> Bucket:
    resource_name = random_suffix_name("s3-bucket", 24)
    resource_data = load_bucket_resource(resource_file_name, resource_name, additional_replacements)

    logging.info(f"Creating bucket {resource_name}")
    # Create k8s resource
    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL,
        resource_name, namespace=namespace,
    )
    resource_data = k8s.create_custom_resource(ref, resource_data)
    k8s.wait_resource_consumed_by_controller(ref)

    time.sleep(CREATE_WAIT_AFTER_SECONDS)

    return Bucket(ref, resource_name, resource_data)

def replace_bucket_spec(bucket: Bucket, resource_file_name: str):
    resource_data = load_bucket_resource(resource_file_name, bucket.resource_name)
    
    # Fetch latest version before patching
    bucket.resource_data = k8s.get_resource(bucket.ref)
    bucket.resource_data["spec"] = resource_data["spec"]
    bucket.resource_data = k8s.replace_custom_resource(bucket.ref, bucket.resource_data)

    time.sleep(MODIFY_WAIT_AFTER_SECONDS)

def delete_bucket(bucket: Bucket):
    if not k8s.get_resource_exists(bucket.ref):
        return
        
    # Delete k8s resource
    _, deleted = k8s.delete_custom_resource(bucket.ref)
    assert deleted is True

    time.sleep(DELETE_WAIT_AFTER_SECONDS)

@pytest.fixture(scope="function")
def basic_bucket(s3_client) -> Generator[Bucket, None, None]:
    bucket = None
    try:
        bucket = create_bucket("bucket")
        assert k8s.get_resource_exists(bucket.ref)
        
        # assert bucket ARN is present in status
        bucket_k8s = bucket.resource_data = k8s.get_resource(bucket.ref)
        assert "arn:aws:s3:::" + bucket.resource_name == bucket_k8s["status"]["ackResourceMetadata"]["arn"]

        exists = bucket_exists(s3_client, bucket)
        assert exists
    except:
        if bucket is not None:
            delete_bucket(bucket)
        return pytest.fail("Bucket failed to create")

    yield bucket

    delete_bucket(bucket)
    exists = bucket_exists(s3_client, bucket)
    assert not exists

@service_marker
class TestBucket:
    def test_basic(self, basic_bucket):
        # Existance assertions are handled by the fixture
        assert basic_bucket

    def test_put_fields(self, s3_client, s3_resource, basic_bucket):
        self._update_assert_accelerate(basic_bucket, s3_client)
        self._update_assert_cors(basic_bucket, s3_resource)
        self._update_assert_encryption(basic_bucket, s3_client)
        self._update_assert_lifecycle(basic_bucket, s3_resource)
        self._update_assert_logging(basic_bucket, s3_resource)
        self._update_assert_notification(basic_bucket, s3_resource)
        self._update_assert_ownership_controls(basic_bucket, s3_client)
        self._update_assert_policy(basic_bucket, s3_resource)
        self._update_assert_public_access_block(basic_bucket, s3_client)
        self._update_assert_replication(basic_bucket, s3_client)
        self._update_assert_request_payment(basic_bucket, s3_resource)
        self._update_assert_tagging(basic_bucket, s3_resource)
        self._update_assert_versioning(basic_bucket, s3_resource)
        self._update_assert_website(basic_bucket, s3_resource)

    def _update_assert_accelerate(self, bucket: Bucket, s3_client):
        replace_bucket_spec(bucket, "bucket_accelerate")

        accelerate_configuration = s3_client.get_bucket_accelerate_configuration(Bucket=bucket.resource_name)

        desired = bucket.resource_data["spec"]["accelerate"]
        latest = accelerate_configuration

        assert desired["status"] == latest["Status"]

    def _update_assert_cors(self, bucket: Bucket, s3_resource):
        replace_bucket_spec(bucket, "bucket_cors")
        
        latest = get_bucket(s3_resource, bucket.resource_name)
        cors = latest.Cors()

        desired_rule = bucket.resource_data["spec"]["cors"]["corsRules"][0]
        latest_rule = cors.cors_rules[0]

        assert desired_rule.get("allowedMethods", []) == latest_rule.get("AllowedMethods", [])
        assert desired_rule.get("allowedOrigins", []) == latest_rule.get("AllowedOrigins", [])
        assert desired_rule.get("allowedHeaders", []) == latest_rule.get("AllowedHeaders", [])
        assert desired_rule.get("exposeHeaders", []) == latest_rule.get("ExposeHeaders", [])

    def _update_assert_encryption(self, bucket: Bucket, s3_client):
        replace_bucket_spec(bucket, "bucket_encryption")

        encryption = s3_client.get_bucket_encryption(Bucket=bucket.resource_name)
        
        desired_rule = bucket.resource_data["spec"]["encryption"]["rules"][0]
        latest_rule = encryption["ServerSideEncryptionConfiguration"]["Rules"][0]

        assert desired_rule["applyServerSideEncryptionByDefault"]["sseAlgorithm"] == \
            latest_rule["ApplyServerSideEncryptionByDefault"]["SSEAlgorithm"]

    def _update_assert_lifecycle(self, bucket: Bucket, s3_resource):
        replace_bucket_spec(bucket, "bucket_lifecycle")

        latest = get_bucket(s3_resource, bucket.resource_name)
        request_payment = latest.LifecycleConfiguration()

        desired_rule = bucket.resource_data["spec"]["lifecycle"]["rules"][0]
        latest_rule = request_payment.rules[0]

        assert desired_rule["id"] == latest_rule["ID"]
        assert desired_rule["status"] == latest_rule["Status"]

    def _update_assert_logging(self, bucket: Bucket, s3_resource):
        replace_bucket_spec(bucket, "bucket_logging")
        
        latest = get_bucket(s3_resource, bucket.resource_name)
        logging = latest.Logging()

        desired = bucket.resource_data["spec"]["logging"]["loggingEnabled"]
        latest = logging.logging_enabled

        assert desired["targetBucket"] == latest["TargetBucket"]
        assert desired["targetPrefix"] == latest["TargetPrefix"]

    def _update_assert_notification(self, bucket: Bucket, s3_resource):
        replace_bucket_spec(bucket, "bucket_notification")
        
        latest = get_bucket(s3_resource, bucket.resource_name)
        notification = latest.Notification()

        desired_config = bucket.resource_data["spec"]["notification"]["topicConfigurations"][0]
        latest_config = notification.topic_configurations[0]

        assert desired_config["id"] == latest_config["Id"]
        assert desired_config["topicARN"] == latest_config["TopicArn"]

    def _update_assert_ownership_controls(self, bucket: Bucket, s3_client):
        replace_bucket_spec(bucket, "bucket_ownership_controls")

        ownership_controls = s3_client.get_bucket_ownership_controls(Bucket=bucket.resource_name)
        
        desired_rule = bucket.resource_data["spec"]["ownershipControls"]["rules"][0]
        latest_rule = ownership_controls["OwnershipControls"]["Rules"][0]

        assert desired_rule["objectOwnership"] == latest_rule["ObjectOwnership"]

    def _update_assert_policy(self, bucket: Bucket, s3_resource):
        replace_bucket_spec(bucket, "bucket_policy")

        latest = get_bucket(s3_resource, bucket.resource_name)
        policy = latest.Policy()

        # Strip any whitespace from between the two
        desired = re.sub(r"\s+", "", bucket.resource_data["spec"]["policy"], flags=re.UNICODE)
        latest = re.sub(r"\s+", "", policy.policy, flags=re.UNICODE)

        assert desired == latest

    def _update_assert_public_access_block(self, bucket: Bucket, s3_client):
        replace_bucket_spec(bucket, "bucket_public_access_block")

        public_access_block = s3_client.get_public_access_block(Bucket=bucket.resource_name)

        desired = bucket.resource_data["spec"]["publicAccessBlock"]
        latest = public_access_block["PublicAccessBlockConfiguration"]

        assert desired["blockPublicACLs"] == latest["BlockPublicAcls"]
        assert desired["blockPublicPolicy"] == latest["BlockPublicPolicy"]

    def _update_assert_replication(self, bucket: Bucket, s3_client):
        replace_bucket_spec(bucket, "bucket_replication")
        
        replication = s3_client.get_bucket_replication(Bucket=bucket.resource_name)

        desired = bucket.resource_data["spec"]["replication"]
        latest = replication["ReplicationConfiguration"]

        desired_rule = desired["rules"][0]
        latest_rule = latest["Rules"][0]

        assert desired["role"] == latest["Role"]
        assert desired_rule["id"] == latest_rule["ID"]
        assert desired_rule["destination"]["bucket"] == latest_rule["Destination"]["Bucket"]

    def _update_assert_request_payment(self, bucket: Bucket, s3_resource):
        replace_bucket_spec(bucket, "bucket_request_payment")

        latest = get_bucket(s3_resource, bucket.resource_name)
        request_payment = latest.RequestPayment()

        desired = bucket.resource_data["spec"]["requestPayment"]["payer"]
        latest = request_payment.payer

        assert desired == latest

    def _update_assert_tagging(self, bucket: Bucket, s3_resource):
        replace_bucket_spec(bucket, "bucket_tagging")

        latest = get_bucket(s3_resource, bucket.resource_name)
        tagging = latest.Tagging()

        desired = bucket.resource_data["spec"]["tagging"]["tagSet"]
        latest = tags.clean(tagging.tag_set)

        for i in range(2):
            assert desired[i]["key"] == latest[i]["Key"]
            assert desired[i]["value"] == latest[i]["Value"]

    def _update_assert_versioning(self, bucket: Bucket, s3_resource):
        replace_bucket_spec(bucket, "bucket_versioning")

        latest = get_bucket(s3_resource, bucket.resource_name)
        versioning = latest.Versioning()

        desired = bucket.resource_data["spec"]["versioning"]["status"]
        latest = versioning.status

        assert desired == latest

    def _update_assert_website(self, bucket: Bucket, s3_resource):
        replace_bucket_spec(bucket, "bucket_website")

        latest = get_bucket(s3_resource, bucket.resource_name)
        website = latest.Website()

        desired = bucket.resource_data["spec"]["website"]
        latest = website

        assert desired["errorDocument"]["key"] == latest.error_document["Key"]
        assert desired["indexDocument"]["suffix"] == latest.index_document["Suffix"]