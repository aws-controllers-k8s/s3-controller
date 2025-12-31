	ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, aws.String("bucket created, requeue for updates"), nil)
	err = ackrequeue.NeededAfter(fmt.Errorf("trigger update"), time.Second)
	return &resource{ko}, err
