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

import pytest
from pathlib import Path
from acktest.k8s import resource

SERVICE_NAME = "s3"
CRD_GROUP = "s3.services.k8s.aws"
CRD_VERSION = "v1alpha1"

# PyTest marker for the current service
service_marker = pytest.mark.service(arg=SERVICE_NAME)

resource_directory = Path(__file__).parent / "resources"

def create_s3_resource(
    resource_plural, resource_name, spec_file, replacements, namespace="default"
):
    """
    Wrapper around k8s.load_and_create_resource to create an S3 resource
    """

    reference, spec, resource = resource.load_and_create_resource(
        SERVICE_NAME,
        CRD_GROUP,
        CRD_VERSION,
        resource_plural,
        resource_name,
        spec_file,
        replacements,
        namespace,
    )

    return reference, spec, resource
