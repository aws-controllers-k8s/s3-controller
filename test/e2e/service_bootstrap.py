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
"""Bootstraps the resources required to run the S3 integration tests.
"""

import logging
import json
from pathlib import Path

from acktest.bootstrapping import Resources, BootstrapFailureException
from acktest.bootstrapping.iam import Role, UserPolicies
from acktest.bootstrapping.s3 import Bucket
from acktest.bootstrapping.sns import Topic
from e2e import bootstrap_directory
from e2e.bootstrap_resources import BootstrapResources


def service_bootstrap() -> Resources:
    logging.getLogger().setLevel(logging.INFO)
    
    replication_policy = json.dumps({
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Action": [
                    "s3:ReplicateObject",
                    "s3:GetObjectVersionTagging",
                    "s3:ReplicateTags",
                    "s3:GetObjectVersionAcl",
                    "s3:ListBucket",
                    "s3:GetReplicationConfiguration",
                    "s3:ReplicateDelete",
                    "s3:GetObjectVersion"
                ],
                "Resource": "*"
            }
        ]
    })

    notification_policy = json.dumps({
        "Version": "2008-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Principal": {
                    "Service": "s3.amazonaws.com"
                },
                "Action": "SNS:Publish",
                "Resource": "*"
            }
        ]
        })

    resources = BootstrapResources(
        ReplicationBucket=Bucket("ack-s3-replication", enable_versioning=True),
        AdoptionBucket=Bucket("ack-s3-annotation-adoption", enable_versioning=True),
        ReplicationRole=Role("ack-s3-replication-role", "s3.amazonaws.com",
            user_policies=UserPolicies("ack-s3-replication-policy", [replication_policy])
        ),
        NotificationTopic=Topic("ack-s3-notification", policy=notification_policy)
    )

    try:
        resources.bootstrap()
    except BootstrapFailureException as ex:
        exit(254)

    return resources

if __name__ == "__main__":
    config = service_bootstrap()
    # Write config to current directory by default
    config.serialize(bootstrap_directory)