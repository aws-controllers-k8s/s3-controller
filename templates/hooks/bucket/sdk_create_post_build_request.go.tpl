
	// Validate directory bucket spec before creation
	if err := validateDirectoryBucketSpec(desired.ko); err != nil {
		return nil, err
	}

	// Only set default LocationConstraint for general-purpose buckets
	// Directory buckets use CreateBucketConfiguration.Location instead
	if desired.ko.Spec.Name == nil || !IsDirectoryBucketName(*desired.ko.Spec.Name) {
		if rm.awsRegion != "us-east-1" {
			// Set default region if not specified
			if input.CreateBucketConfiguration == nil ||
				input.CreateBucketConfiguration.LocationConstraint == "" {
				input.CreateBucketConfiguration = &svcsdktypes.CreateBucketConfiguration{
					LocationConstraint: svcsdktypes.BucketLocationConstraint(rm.awsRegion),
				}
			}
		}
	}

	// Copy Spec.Tagging into the request's CreateBucketConfiguration.Tags so
	// that tags are applied at creation time. This allows bucket creation to
	// succeed under IAM/SCP policies that enforce mandatory tags on
	// s3:CreateBucket via aws:RequestTag conditions. Spec.Tagging remains the
	// source of truth; post-create tag updates still flow through
	// PutBucketTagging. NOTE: this must run after the LocationConstraint
	// defaulting above, which may replace input.CreateBucketConfiguration.
	addCreateBucketTags(desired.ko, input)
