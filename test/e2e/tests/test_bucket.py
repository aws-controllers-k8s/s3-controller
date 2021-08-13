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

from acktest.resources import random_suffix_name
from acktest.k8s import resource as k8s
from e2e import service_marker, CRD_GROUP, CRD_VERSION, load_s3_resource
from e2e.replacement_values import REPLACEMENT_VALUES
from e2e.bootstrap_resources import TestBootstrapResources, get_bootstrap_resources

RESOURCE_PLURAL = "buckets"

CREATE_WAIT_AFTER_SECONDS = 10
DELETE_WAIT_AFTER_SECONDS = 10

def get_bucket(s3_resource, bucket_name: str):
    return s3_resource.Bucket(bucket_name)

def bucket_exists(s3_client, bucket_name: str) -> bool:
    try:
        resp = s3_client.list_buckets()
    except Exception as e:
        logging.debug(e)
        return False

    buckets = resp["Buckets"]
    for bucket in buckets:
        if bucket["Name"] == bucket_name:
            return True

    return False

def create_bucket(s3_client, resource_file_name: str):
    resource_name = random_suffix_name("s3-bucket", 24)

    replacements = REPLACEMENT_VALUES.copy()
    replacements["BUCKET_NAME"] = resource_name

    # Load Bucket CR
    resource_data = load_s3_resource(
        resource_file_name,
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
    exists = bucket_exists(s3_client, resource_name)
    assert exists

    return (ref, resource_name, resource_data)

def delete_bucket(s3_client, ref, resource_name):
    # Delete k8s resource
    _, deleted = k8s.delete_custom_resource(ref)
    assert deleted is True

    time.sleep(DELETE_WAIT_AFTER_SECONDS)

    # Check S3 Bucket doesn't exists
    exists = bucket_exists(s3_client, resource_name)
    assert not exists

@pytest.fixture(scope="module")
def s3_client():
    return boto3.client("s3")

@pytest.fixture(scope="module")
def s3_resource():
    return boto3.resource("s3")

@service_marker
@pytest.mark.canary
class TestBucket:
    def test_basic(self, s3_client):
        (ref, resource_name, _) = create_bucket(s3_client, "bucket")
        delete_bucket(s3_client, ref, resource_name)

    def test_accelerate(self, s3_client):
        (ref, resource_name, resource_data) = create_bucket(s3_client, "bucket_accelerate")
        
        accelerate_configuration = s3_client.get_bucket_accelerate_configuration(Bucket=resource_name)
        assert resource_data["spec"]["accelerate"]["status"] == accelerate_configuration["Status"]

        delete_bucket(s3_client, ref, resource_name)

    def test_cors(self, s3_client, s3_resource):
        (ref, resource_name, resource_data) = create_bucket(s3_client, "bucket_cors")
        
        # Do checks
        bucket = get_bucket(s3_resource, resource_name)
        cors = bucket.Cors()

        desired_rule = resource_data["spec"]["cors"]["corsRules"][0]
        latest_rule = cors.cors_rules[0]

        assert desired_rule.get("allowedMethods", []) == latest_rule.get("AllowedMethods", [])
        assert desired_rule.get("allowedOrigins", []) == latest_rule.get("AllowedOrigins", [])
        assert desired_rule.get("allowedHeaders", []) == latest_rule.get("AllowedHeaders", [])
        assert desired_rule.get("exposeHeaders", []) == latest_rule.get("ExposeHeaders", [])

        delete_bucket(s3_client, ref, resource_name)

    def test_encryption(self, s3_client):
        (ref, resource_name, resource_data) = create_bucket(s3_client, "bucket_encryption")

        encryption = s3_client.get_bucket_encryption(Bucket=resource_name)
        
        desired_rule = resource_data["spec"]["encryption"]["rules"][0]
        latest_rule = encryption["ServerSideEncryptionConfiguration"]["Rules"][0]

        assert desired_rule["applyServerSideEncryptionByDefault"]["sseAlgorithm"] == \
            latest_rule["ApplyServerSideEncryptionByDefault"]["SSEAlgorithm"]

        delete_bucket(s3_client, ref, resource_name)

    def test_logging(self, s3_client, s3_resource):
        (ref, resource_name, resource_data) = create_bucket(s3_client, "bucket_logging")
        
        bucket = get_bucket(s3_resource, resource_name)
        logging = bucket.Logging()

        desired = resource_data["spec"]["logging"]["loggingEnabled"]
        latest = logging.logging_enabled

        assert desired["targetBucket"] == latest["TargetBucket"]
        assert desired["targetPrefix"] == latest["TargetPrefix"]

        delete_bucket(s3_client, ref, resource_name)

    def test_ownership_controls(self, s3_client):
        (ref, resource_name, resource_data) = create_bucket(s3_client, "bucket_ownership_controls")

        ownership_controls = s3_client.get_bucket_ownership_controls(Bucket=resource_name)
        
        desired_rule = resource_data["spec"]["ownershipControls"]["rules"][0]
        latest_rule = ownership_controls["OwnershipControls"]["Rules"][0]

        assert desired_rule["objectOwnership"] == latest_rule["ObjectOwnership"]

        delete_bucket(s3_client, ref, resource_name)

    def test_request_payment(self, s3_client, s3_resource):
        (ref, resource_name, resource_data) = create_bucket(s3_client, "bucket_request_payment")

        bucket = get_bucket(s3_resource, resource_name)
        request_payment = bucket.RequestPayment()

        desired = resource_data["spec"]["requestPayment"]["payer"]
        latest = request_payment.payer

        assert desired == latest

        delete_bucket(s3_client, ref, resource_name)

    def test_tagging(self, s3_client, s3_resource):
        (ref, resource_name, resource_data) = create_bucket(s3_client, "bucket_tagging")

        bucket = get_bucket(s3_resource, resource_name)
        tagging = bucket.Tagging()

        desired = resource_data["spec"]["tagging"]["tagSet"]
        latest = tagging.tag_set

        for i in range(2):
            assert desired[i]["key"] == latest[i]["Key"]
            assert desired[i]["value"] == latest[i]["Value"]

        delete_bucket(s3_client, ref, resource_name)

    def test_versioning(self, s3_client, s3_resource):
        (ref, resource_name, resource_data) = create_bucket(s3_client, "bucket_versioning")

        bucket = get_bucket(s3_resource, resource_name)
        versioning = bucket.Versioning()

        desired = resource_data["spec"]["versioning"]["status"]
        latest = versioning.status

        assert desired == latest

        delete_bucket(s3_client, ref, resource_name)

    def test_website(self, s3_client, s3_resource):
        (ref, resource_name, resource_data) = create_bucket(s3_client, "bucket_website")

        bucket = get_bucket(s3_resource, resource_name)
        website = bucket.Website()

        desired = resource_data["spec"]["website"]
        latest = website

        assert desired["errorDocument"]["key"] == latest.error_document["Key"]
        assert desired["indexDocument"]["suffix"] == latest.index_document["Suffix"]

        delete_bucket(s3_client, ref, resource_name)