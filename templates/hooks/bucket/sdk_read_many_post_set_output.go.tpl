    // Describe and set bucket logging
    getBucketLoggingPayload := rm.newGetBucketLoggingPayload(r)
	getBucketLoggingResponse, err := rm.sdkapi.GetBucketLoggingWithContext(ctx, getBucketLoggingPayload)
	if err != nil {
		return nil, err
	}
	ko.Spec.Logging = rm.setResourceLogging(r, getBucketLoggingResponse)