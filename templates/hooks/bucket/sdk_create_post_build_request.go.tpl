	if input.CreateBucketConfiguration == nil {
		input.CreateBucketConfiguration = &svcsdktypes.CreateBucketConfiguration{}
	}
	if input.CreateBucketConfiguration.LocationConstraint == "" && rm.awsRegion != "us-east-1" {
		input.CreateBucketConfiguration.LocationConstraint = svcsdktypes.BucketLocationConstraint(rm.awsRegion)
	}