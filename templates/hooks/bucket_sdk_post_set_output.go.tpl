    latest := &resource{ko}
	synced, err := rm.syncPutFields(ctx, latest)
	if err != nil {
		return nil, err
	}
	ko = synced.ko