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
from acktest.k8s import resource as k8s, condition
from acktest.aws.identity import get_region, get_account_id
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

def get_availability_zone() -> tuple[str, str]:
    """Retrieve an available availability zone for directory bucket creation.
    
    Returns:
        tuple[str, str]: A tuple containing (zone_name, zone_id) where:
            - zone_name: The name of the availability zone (e.g., 'us-west-2a')
            - zone_id: The ID of the availability zone (e.g., 'usw2-az1')
        
    Raises:
        Exception: If no availability zones are available or if the EC2 API call fails.
    """
    try:
        region = get_region()
        ec2_client = boto3.client('ec2', region_name=region)
        
        response = ec2_client.describe_availability_zones(
            Filters=[
                {
                    'Name': 'state',
                    'Values': ['available']
                }
            ]
        )
        
        availability_zones = response.get('AvailabilityZones', [])
        
        if not availability_zones:
            raise Exception(f"No available availability zones found in region {region}")
        
        # Return the first available zone's name and ID
        az_name = availability_zones[0]['ZoneName']
        az_id = availability_zones[0]['ZoneId']
        return az_name, az_id
        
    except Exception as e:
        logging.error(f"Failed to retrieve availability zone: {e}")
        raise

def generate_directory_bucket_name(base_name: str, az_id: str) -> str:
    """Generate a valid directory bucket name with the required suffix.
    
    Directory buckets must end with '--azid--x-s3' where azid is the availability zone ID.
    This function takes a base name and appends the proper suffix.
    
    Args:
        base_name: The base name for the bucket (e.g., 's3-bucket-abc123')
        az_id: The availability zone ID (e.g., 'usw2-az1')
        
    Returns:
        str: A valid directory bucket name with the proper suffix
        
    Example:
        >>> generate_directory_bucket_name('my-bucket-abc123', 'usw2-az1')
        'my-bucket-abc123--usw2-az1--x-s3'
    """
    if not base_name:
        raise ValueError("base_name cannot be empty")
    if not az_id:
        raise ValueError("az_id cannot be empty")
    
    # Construct the directory bucket name with the required suffix
    directory_bucket_name = f"{base_name}--{az_id}--x-s3"
    
    logging.info(f"Generated directory bucket name: {directory_bucket_name}")
    return directory_bucket_name

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

def directory_bucket_exists(s3_client, bucket: Bucket) -> bool:
    """Check if a directory bucket exists in AWS.
    
    Directory buckets are identified by their BucketType field in the list_buckets response.
    This function filters the bucket list to only directory buckets before checking for existence.
    
    Args:
        s3_client: Boto3 S3 client instance
        bucket: Bucket dataclass instance containing the bucket name to check
        
    Returns:
        bool: True if the directory bucket exists, False otherwise
    """
    try:
        resp = s3_client.list_directory_buckets()
    except Exception as e:
        logging.error(f"Failed to list buckets: {e}")
        return False

    buckets = resp.get("Buckets", [])
    for _bucket in buckets:
        if _bucket["Name"] == bucket.resource_name:
            return True

    return False

def load_bucket_resource(resource_file_name: str, resource_name: str, additional_replacements: dict = None):
    additional_replacements = {} if additional_replacements is None else additional_replacements
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

def create_directory_bucket(resource_file_name: str = "directory_bucket", namespace: str = "default") -> Bucket:
    az_name, az_id = get_availability_zone()
    base_name = random_suffix_name("s3-bucket", 24)
    directory_bucket_name = generate_directory_bucket_name(base_name, az_id)
    
    additional_replacements = {
        "LOCATION_NAME": az_id,
        "BUCKET_NAME": directory_bucket_name
    }
    
    replacements = {**REPLACEMENT_VALUES.copy(), **additional_replacements}
    resource_data = load_s3_resource(
        resource_file_name,
        additional_replacements=replacements,
    )
    logging.debug(resource_data)
    
    logging.info(f"Creating directory bucket {directory_bucket_name}")
    # Create k8s resource
    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL,
        directory_bucket_name, namespace=namespace,
    )
    resource_data = k8s.create_custom_resource(ref, resource_data)
    k8s.wait_resource_consumed_by_controller(ref)

    time.sleep(CREATE_WAIT_AFTER_SECONDS)
    assert k8s.wait_on_condition(ref, condition.CONDITION_TYPE_RESOURCE_SYNCED, "True", wait_periods=10)
    
    bucket = Bucket(ref, directory_bucket_name, resource_data)
    logging.info(f"Created directory bucket {directory_bucket_name} in availability zone {az_name}")
    return bucket

def replace_bucket_spec(bucket: Bucket, resource_file_name: str, additional_replacements: dict = None):
    resource_data = load_bucket_resource(resource_file_name, bucket.resource_name, additional_replacements)
    
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

@pytest.fixture(scope="function")
def directory_bucket(s3_client) -> Generator[Bucket, None, None]:
    bucket = None
    try:
        bucket = create_directory_bucket()
        assert k8s.get_resource_exists(bucket.ref)

        exists = directory_bucket_exists(s3_client, bucket)
        assert exists
    except:
        if bucket is not None:
            delete_bucket(bucket)
        return pytest.fail("Directory bucket failed to create")

    yield bucket

    delete_bucket(bucket)
    exists = directory_bucket_exists(s3_client, bucket)
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
        self._update_assert_empty_policy(basic_bucket, s3_resource)
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
    
    def _update_assert_directory_lifecycle(self, bucket: Bucket, s3_resource):
        replace_bucket_spec(bucket, "directory_bucket_lifecycle")

        latest = get_bucket(s3_resource, bucket.resource_name)
        lifecycleConfig = latest.LifecycleConfiguration()

        assert len(lifecycleConfig.rules) == len(bucket.resource_data["spec"]["lifecycle"]["rules"])

        desired_rule = bucket.resource_data["spec"]["lifecycle"]["rules"][0]
        latest_rule = lifecycleConfig.rules[0]

        assert desired_rule["id"] == latest_rule["ID"]
        assert desired_rule["status"] == latest_rule["Status"]
        assert desired_rule["filter"]["prefix"] == latest_rule["Filter"]["Prefix"]
        assert desired_rule["abortIncompleteMultipartUpload"]["daysAfterInitiation"] == latest_rule["AbortIncompleteMultipartUpload"]["DaysAfterInitiation"]

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

    def _update_assert_empty_policy(self, bucket: Bucket, s3_resource):
        replace_bucket_spec(bucket, "bucket_empty_policy")

        latest = get_bucket(s3_resource, bucket.resource_name)
        policy = latest.Policy()

        try:
            policy.policy
        except Exception as e:
            assert "NoSuchBucketPolicy" in str(e)
            return
        
        # if there is no error, fail test
        assert False

    def _update_assert_directory_policy(self, bucket: Bucket, s3_resource):
        arn = k8s.get_resource_arn(bucket.resource_data)
        additional_replacements = {
            "BUCKET_ARN": arn
        }
        replace_bucket_spec(bucket, "directory_bucket_policy", additional_replacements)

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

    def _update_assert_tagging_directory_bucket(self, bucket: Bucket, s3control_client):
        replace_bucket_spec(bucket, "bucket_tagging")
        account_id = str(get_account_id())
        arn = k8s.get_resource_arn(bucket.resource_data)

        aws_tags = s3control_client.list_tags_for_resource(AccountId=account_id, ResourceArn=arn)["Tags"]
        desired = bucket.resource_data["spec"]["tagging"]["tagSet"]
        latest = tags.clean(aws_tags)

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

    def test_directory_bucket_put_fields(self, directory_bucket, s3_client, s3_resource, s3control_client):
        """Test tag updates on directory buckets."""
        self._update_assert_encryption(directory_bucket, s3_client)
        self._update_assert_directory_lifecycle(directory_bucket, s3_resource)
        self._update_assert_tagging_directory_bucket(directory_bucket, s3control_client)
        self._update_assert_directory_policy(directory_bucket, s3_resource)

