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
import logging
import itertools
from typing import TYPE_CHECKING, Generator, List, NamedTuple, Tuple

from acktest.resources import random_suffix_name
from acktest.k8s import resource as k8s
from acktest import adoption as adoption
from acktest import tags as tags
from e2e import SERVICE_NAME
from e2e.tests.test_bucket import Bucket, bucket_exists, create_bucket, delete_bucket

DELETION_POLICY_RESOURCE_ANNOTATION_KEY = "services.k8s.aws/deletion-policy"
DELETION_POLICY_NAMESPACE_ANNOTATION_KEY = (
    f"{SERVICE_NAME}.services.k8s.aws/deletion-policy"
)

class DeletionPolicy(str, Enum):
    NONE = ""
    DELETE = "delete"
    RETAIN = "retain"


# DeletionPolicyAnnotationTuple represents a tuple of namespace and resource
# deletion policy annotations. These are used when testing all combinations of
# each annotation.
class DeletionPolicyAnnotationTuple(NamedTuple):
    namespace: DeletionPolicy
    resource: DeletionPolicy


# Create a matrix of combinations for deletion policy annotations that can be
# used as a parameter for tests
DELETION_POLICY_ANNOTATION_COMBINATIONS: List[DeletionPolicyAnnotationTuple] = [
    DeletionPolicyAnnotationTuple(r[0], r[1])
    for r in itertools.product([p for p in DeletionPolicy], [p for p in DeletionPolicy])
]

# Parameter types are not support by pytest. This adds support for type checking
# the param type.
if TYPE_CHECKING:

    class DeletionPolicyFixtureRequest:
        param: DeletionPolicyAnnotationTuple

else:
    from typing import Any

    DeletionPolicyFixtureRequest = Any


def create_deletion_policy_namespace(deletion_policy: DeletionPolicy) -> str:
    namespace_name = random_suffix_name("s3-deletion-policy", 24)
    annotations = {}
    if deletion_policy != DeletionPolicy.NONE:
        annotations[DELETION_POLICY_NAMESPACE_ANNOTATION_KEY] = deletion_policy.value

    logging.info(f"Creating namespace {namespace_name}")
    try:
        k8s.create_k8s_namespace(namespace_name, annotations)
    except Exception as ex:
        return pytest.fail("Failed to create namespace")

    return namespace_name


@pytest.fixture(scope="function")
def deletion_policy_namespace_bucket(
    request: DeletionPolicyFixtureRequest, s3_client
) -> Generator[Tuple[Bucket, DeletionPolicyAnnotationTuple], None, None]:
    bucket_namespace = create_deletion_policy_namespace(request.param.namespace)

    bucket = None
    try:
        if request.param.resource == DeletionPolicy.NONE:
            bucket = create_bucket("bucket", namespace=bucket_namespace)
        else:
            bucket = create_bucket(
                "bucket_deletion_policy",
                namespace=bucket_namespace,
                additional_replacements={
                    "DELETION_POLICY": request.param.resource.value
                },
            )

        assert k8s.get_resource_exists(bucket.ref)

        exists = bucket_exists(s3_client, bucket)
        assert exists
    except:
        if bucket is not None:
            delete_bucket(bucket)
        return pytest.fail("Bucket failed to create")

    yield (bucket, request.param)

    delete_bucket(bucket)

    exists = bucket_exists(s3_client, bucket)
    if exists:
        s3_client.delete_bucket(Bucket=bucket.resource_name)

    k8s.delete_k8s_namespace(bucket_namespace)


class TestDeletionPolicyBucket:
    @pytest.mark.parametrize(
        "deletion_policy_namespace_bucket",
        DELETION_POLICY_ANNOTATION_COMBINATIONS,
        indirect=True,
    )
    def test_deletion_policy(
        self, s3_client, deletion_policy_namespace_bucket
    ):
        (bucket, deletion_policy_annotations) = deletion_policy_namespace_bucket

        delete_bucket(bucket)

        exists = bucket_exists(s3_client, bucket)

        # Assert in order of precedence (resource > namespace)
        if deletion_policy_annotations.resource == DeletionPolicy.DELETE:
            assert not exists
        elif deletion_policy_annotations.resource == DeletionPolicy.RETAIN:
            assert exists
        elif deletion_policy_annotations.namespace == DeletionPolicy.DELETE:
            assert not exists
        elif deletion_policy_annotations.namespace == DeletionPolicy.RETAIN:
            assert exists
        else: # Neither has an annotation
            assert not exists
