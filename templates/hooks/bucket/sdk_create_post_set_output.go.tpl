	ackcondition.SetSynced(&resource{ko}, corev1.ConditionFalse, aws.String("bucket created, requeue for updates"), nil)
	err = ackrequeue.NeededAfter(fmt.Errorf("Reconciling to sync additional fields"), time.Second)
	return &resource{ko}, err
