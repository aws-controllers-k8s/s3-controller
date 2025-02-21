
	if rm.awsRegion != "us-east-1" {
		// Set default region if not specified
		if input.CreateBucketConfiguration == nil ||
			input.CreateBucketConfiguration.LocationConstraint == "" {
			input.CreateBucketConfiguration = &svcsdktypes.CreateBucketConfiguration{
				LocationConstraint: svcsdktypes.BucketLocationConstraint(rm.awsRegion),
			}
		}
	}
