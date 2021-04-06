# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
#	 http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.
"""Cleans up the resources created by the S3 bootstrapping process.
"""

import logging
from pathlib import Path

from acktest import resources
from e2e.bootstrap_resources import TestBootstrapResources

def service_cleanup(config: dict):
    logging.getLogger().setLevel(logging.INFO)

    resources = TestBootstrapResources(
        **config
    )

if __name__ == "__main__":
    root_test_path = Path(__file__).parent
    
    bootstrap_config = resources.read_bootstrap_config(root_test_path)
    service_cleanup(bootstrap_config) 