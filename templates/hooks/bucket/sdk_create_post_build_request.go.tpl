
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
