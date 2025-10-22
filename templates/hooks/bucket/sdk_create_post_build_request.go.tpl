	// Set default region for general-purpose buckets only (not directory buckets)
	// Directory buckets use Location/Bucket fields instead of LocationConstraint
	isDirectoryBucket := input.CreateBucketConfiguration != nil &&
		input.CreateBucketConfiguration.Location != nil &&
		input.CreateBucketConfiguration.Bucket != nil

	if rm.awsRegion != "us-east-1" && !isDirectoryBucket {
		// Set default region if not specified
		if input.CreateBucketConfiguration == nil ||
			input.CreateBucketConfiguration.LocationConstraint == "" {
			input.CreateBucketConfiguration = &svcsdktypes.CreateBucketConfiguration{
				LocationConstraint: svcsdktypes.BucketLocationConstraint(rm.awsRegion),
			}
		}
	}
