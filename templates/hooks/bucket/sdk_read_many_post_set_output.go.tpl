	if err := rm.addPutFieldsToSpec(ctx, r, ko); err != nil {
		return nil, err
	}

	// Set bucket ARN in the output
	bucketARN := ackv1alpha1.AWSResourceName(rm.ARNFromName(*ko.Spec.Name))
	ko.Status.ACKResourceMetadata.ARN = &bucketARN