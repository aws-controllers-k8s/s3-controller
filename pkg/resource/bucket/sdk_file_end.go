package bucket

import (
	"context"
	"reflect"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcapitypes "github.com/aws-controllers-k8s/s3-controller/apis/v1alpha1"
	"github.com/aws/aws-sdk-go-v2/aws"
	svcsdk "github.com/aws/aws-sdk-go-v2/service/s3"
	svcsdktypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newAccelerateConfiguration returns a AccelerateConfiguration object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newAccelerateConfiguration(
	r *resource,
) *svcsdktypes.AccelerateConfiguration {
	res := &svcsdktypes.AccelerateConfiguration{}

	if r.ko.Spec.Accelerate.Status != nil {
		res.Status = svcsdktypes.BucketAccelerateStatus(*r.ko.Spec.Accelerate.Status)
	}

	return res
}

// setResourceAccelerate sets the `Accelerate` spec field
// given the output of a `GetBucketAccelerateConfiguration` operation.
func (rm *resourceManager) setResourceAccelerate(
	r *resource,
	resp *svcsdk.GetBucketAccelerateConfigurationOutput,
) *svcapitypes.AccelerateConfiguration {
	res := &svcapitypes.AccelerateConfiguration{}

	if resp.Status != "" {
		res.Status = aws.String(string(resp.Status))
	}

	return res
}

// newAnalyticsConfiguration returns a AnalyticsConfiguration object
// with each the field set by the corresponding configuration's fields.
func (rm *resourceManager) newAnalyticsConfiguration(
	c *svcapitypes.AnalyticsConfiguration,
) *svcsdktypes.AnalyticsConfiguration {
	res := &svcsdktypes.AnalyticsConfiguration{}

	if c.Filter != nil {
		if c.Filter.And != nil {
			resf0 := &svcsdktypes.AnalyticsFilterMemberAnd{}
			resf0f0 := svcsdktypes.AnalyticsAndOperator{}
			if c.Filter.And.Prefix != nil {
				resf0f0.Prefix = c.Filter.And.Prefix
			}
			if c.Filter.And.Tags != nil {
				resf0f0f1 := []svcsdktypes.Tag{}
				for _, resf0f0f1iter := range c.Filter.And.Tags {
					resf0f0f1elem := &svcsdktypes.Tag{}
					if resf0f0f1iter.Key != nil {
						resf0f0f1elem.Key = resf0f0f1iter.Key
					}
					if resf0f0f1iter.Value != nil {
						resf0f0f1elem.Value = resf0f0f1iter.Value
					}
					resf0f0f1 = append(resf0f0f1, *resf0f0f1elem)
				}
				resf0f0.Tags = resf0f0f1
			}
			resf0.Value = resf0f0
			res.Filter = resf0
		}
		if c.Filter.Prefix != nil {
			resf0 := &svcsdktypes.AnalyticsFilterMemberPrefix{}
			resf0.Value = *c.Filter.Prefix
			res.Filter = resf0
		}
		if c.Filter.Tag != nil {
			resf0 := &svcsdktypes.AnalyticsFilterMemberTag{}
			resf0f2 := svcsdktypes.Tag{}
			if c.Filter.Tag.Key != nil {
				resf0f2.Key = c.Filter.Tag.Key
			}
			if c.Filter.Tag.Value != nil {
				resf0f2.Value = c.Filter.Tag.Value
			}
			resf0.Value = resf0f2
			res.Filter = resf0
		}
	}
	if c.ID != nil {
		res.Id = c.ID
	}
	if c.StorageClassAnalysis != nil {
		resf2 := &svcsdktypes.StorageClassAnalysis{}
		if c.StorageClassAnalysis.DataExport != nil {
			resf2f0 := &svcsdktypes.StorageClassAnalysisDataExport{}
			if c.StorageClassAnalysis.DataExport.Destination != nil {
				resf2f0f0 := &svcsdktypes.AnalyticsExportDestination{}
				if c.StorageClassAnalysis.DataExport.Destination.S3BucketDestination != nil {
					resf2f0f0f0 := &svcsdktypes.AnalyticsS3BucketDestination{}
					if c.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Bucket != nil {
						resf2f0f0f0.Bucket = c.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Bucket
					}
					if c.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.BucketAccountID != nil {
						resf2f0f0f0.BucketAccountId = c.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.BucketAccountID
					}
					if c.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Format != nil {
						resf2f0f0f0.Format = svcsdktypes.AnalyticsS3ExportFileFormat(*c.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Format)
					}
					if c.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Prefix != nil {
						resf2f0f0f0.Prefix = c.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Prefix
					}
					resf2f0f0.S3BucketDestination = resf2f0f0f0
				}
				resf2f0.Destination = resf2f0f0
			}
			if c.StorageClassAnalysis.DataExport.OutputSchemaVersion != nil {
				resf2f0.OutputSchemaVersion = svcsdktypes.StorageClassAnalysisSchemaVersion(*c.StorageClassAnalysis.DataExport.OutputSchemaVersion)
			}
			resf2.DataExport = resf2f0
		}
		res.StorageClassAnalysis = resf2
	}

	return res
}

// setAnalyticsConfiguration sets a resource AnalyticsConfiguration type
// given the SDK type.
func (rm *resourceManager) setResourceAnalyticsConfiguration(
	r *resource,
	resp svcsdktypes.AnalyticsConfiguration,
) *svcapitypes.AnalyticsConfiguration {
	res := &svcapitypes.AnalyticsConfiguration{}

	if resp.Filter != nil {
		resf0 := &svcapitypes.AnalyticsFilter{}
		switch resp.Filter.(type) {
		case *svcsdktypes.AnalyticsFilterMemberAnd:
			f0 := resp.Filter.(*svcsdktypes.AnalyticsFilterMemberAnd)
			resf0f0 := &svcapitypes.AnalyticsAndOperator{}
			if f0.Value.Prefix != nil {
				resf0f0.Prefix = f0.Value.Prefix
			}
			if f0.Value.Tags != nil {
				resf0f0f1 := []*svcapitypes.Tag{}
				for _, resf0f0f1iter := range f0.Value.Tags {
					resf0f0f1elem := &svcapitypes.Tag{}
					if resf0f0f1iter.Key != nil {
						resf0f0f1elem.Key = resf0f0f1iter.Key
					}
					if resf0f0f1iter.Value != nil {
						resf0f0f1elem.Value = resf0f0f1iter.Value
					}
					resf0f0f1 = append(resf0f0f1, resf0f0f1elem)
				}
				resf0f0.Tags = resf0f0f1
				resf0.And = resf0f0
			}
		case *svcsdktypes.AnalyticsFilterMemberPrefix:
			f0 := resp.Filter.(*svcsdktypes.AnalyticsFilterMemberPrefix)
			resf0.Prefix = &f0.Value
		case *svcsdktypes.AnalyticsFilterMemberTag:
			f0 := resp.Filter.(*svcsdktypes.AnalyticsFilterMemberTag)

			resf0f2 := &svcapitypes.Tag{}
			if f0.Value.Key != nil {
				resf0f2.Key = f0.Value.Key
			}
			if f0.Value.Value != nil {
				resf0f2.Value = f0.Value.Value
			}
			resf0.Tag = resf0f2
		}
		res.Filter = resf0
	}
	if resp.Id != nil {
		res.ID = resp.Id
	}
	if resp.StorageClassAnalysis != nil {
		resf2 := &svcapitypes.StorageClassAnalysis{}
		if resp.StorageClassAnalysis.DataExport != nil {
			resf2f0 := &svcapitypes.StorageClassAnalysisDataExport{}
			if resp.StorageClassAnalysis.DataExport.Destination != nil {
				resf2f0f0 := &svcapitypes.AnalyticsExportDestination{}
				if resp.StorageClassAnalysis.DataExport.Destination.S3BucketDestination != nil {
					resf2f0f0f0 := &svcapitypes.AnalyticsS3BucketDestination{}
					if resp.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Bucket != nil {
						resf2f0f0f0.Bucket = resp.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Bucket
					}
					if resp.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.BucketAccountId != nil {
						resf2f0f0f0.BucketAccountID = resp.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.BucketAccountId
					}
					if resp.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Format != "" {
						resf2f0f0f0.Format = aws.String(string(resp.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Format))
					}
					if resp.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Prefix != nil {
						resf2f0f0f0.Prefix = resp.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Prefix
					}
					resf2f0f0.S3BucketDestination = resf2f0f0f0
				}
				resf2f0.Destination = resf2f0f0
			}
			if resp.StorageClassAnalysis.DataExport.OutputSchemaVersion != "" {
				resf2f0.OutputSchemaVersion = aws.String(string(resp.StorageClassAnalysis.DataExport.OutputSchemaVersion))
			}
			resf2.DataExport = resf2f0
		}
		res.StorageClassAnalysis = resf2
	}

	return res
}

func compareAnalyticsConfiguration(
	a *svcapitypes.AnalyticsConfiguration,
	b *svcapitypes.AnalyticsConfiguration,
) *ackcompare.Delta {
	delta := ackcompare.NewDelta()
	if ackcompare.HasNilDifference(a.Filter, b.Filter) {
		delta.Add("AnalyticsConfiguration.Filter", a.Filter, b.Filter)
	} else if a.Filter != nil && b.Filter != nil {
		if ackcompare.HasNilDifference(a.Filter.And, b.Filter.And) {
			delta.Add("AnalyticsConfiguration.Filter.And", a.Filter.And, b.Filter.And)
		} else if a.Filter.And != nil && b.Filter.And != nil {
			if ackcompare.HasNilDifference(a.Filter.And.Prefix, b.Filter.And.Prefix) {
				delta.Add("AnalyticsConfiguration.Filter.And.Prefix", a.Filter.And.Prefix, b.Filter.And.Prefix)
			} else if a.Filter.And.Prefix != nil && b.Filter.And.Prefix != nil {
				if *a.Filter.And.Prefix != *b.Filter.And.Prefix {
					delta.Add("AnalyticsConfiguration.Filter.And.Prefix", a.Filter.And.Prefix, b.Filter.And.Prefix)
				}
			}
			if len(a.Filter.And.Tags) != len(b.Filter.And.Tags) {
				delta.Add("AnalyticsConfiguration.Filter.And.Tags", a.Filter.And.Tags, b.Filter.And.Tags)
			} else if len(a.Filter.And.Tags) > 0 {
				if !reflect.DeepEqual(a.Filter.And.Tags, b.Filter.And.Tags) {
					delta.Add("AnalyticsConfiguration.Filter.And.Tags", a.Filter.And.Tags, b.Filter.And.Tags)
				}
			}
		}
		if ackcompare.HasNilDifference(a.Filter.Prefix, b.Filter.Prefix) {
			delta.Add("AnalyticsConfiguration.Filter.Prefix", a.Filter.Prefix, b.Filter.Prefix)
		} else if a.Filter.Prefix != nil && b.Filter.Prefix != nil {
			if *a.Filter.Prefix != *b.Filter.Prefix {
				delta.Add("AnalyticsConfiguration.Filter.Prefix", a.Filter.Prefix, b.Filter.Prefix)
			}
		}
		if ackcompare.HasNilDifference(a.Filter.Tag, b.Filter.Tag) {
			delta.Add("AnalyticsConfiguration.Filter.Tag", a.Filter.Tag, b.Filter.Tag)
		} else if a.Filter.Tag != nil && b.Filter.Tag != nil {
			if ackcompare.HasNilDifference(a.Filter.Tag.Key, b.Filter.Tag.Key) {
				delta.Add("AnalyticsConfiguration.Filter.Tag.Key", a.Filter.Tag.Key, b.Filter.Tag.Key)
			} else if a.Filter.Tag.Key != nil && b.Filter.Tag.Key != nil {
				if *a.Filter.Tag.Key != *b.Filter.Tag.Key {
					delta.Add("AnalyticsConfiguration.Filter.Tag.Key", a.Filter.Tag.Key, b.Filter.Tag.Key)
				}
			}
			if ackcompare.HasNilDifference(a.Filter.Tag.Value, b.Filter.Tag.Value) {
				delta.Add("AnalyticsConfiguration.Filter.Tag.Value", a.Filter.Tag.Value, b.Filter.Tag.Value)
			} else if a.Filter.Tag.Value != nil && b.Filter.Tag.Value != nil {
				if *a.Filter.Tag.Value != *b.Filter.Tag.Value {
					delta.Add("AnalyticsConfiguration.Filter.Tag.Value", a.Filter.Tag.Value, b.Filter.Tag.Value)
				}
			}
		}
	}
	if ackcompare.HasNilDifference(a.ID, b.ID) {
		delta.Add("AnalyticsConfiguration.ID", a.ID, b.ID)
	} else if a.ID != nil && b.ID != nil {
		if *a.ID != *b.ID {
			delta.Add("AnalyticsConfiguration.ID", a.ID, b.ID)
		}
	}
	if ackcompare.HasNilDifference(a.StorageClassAnalysis, b.StorageClassAnalysis) {
		delta.Add("AnalyticsConfiguration.StorageClassAnalysis", a.StorageClassAnalysis, b.StorageClassAnalysis)
	} else if a.StorageClassAnalysis != nil && b.StorageClassAnalysis != nil {
		if ackcompare.HasNilDifference(a.StorageClassAnalysis.DataExport, b.StorageClassAnalysis.DataExport) {
			delta.Add("AnalyticsConfiguration.StorageClassAnalysis.DataExport", a.StorageClassAnalysis.DataExport, b.StorageClassAnalysis.DataExport)
		} else if a.StorageClassAnalysis.DataExport != nil && b.StorageClassAnalysis.DataExport != nil {
			if ackcompare.HasNilDifference(a.StorageClassAnalysis.DataExport.Destination, b.StorageClassAnalysis.DataExport.Destination) {
				delta.Add("AnalyticsConfiguration.StorageClassAnalysis.DataExport.Destination", a.StorageClassAnalysis.DataExport.Destination, b.StorageClassAnalysis.DataExport.Destination)
			} else if a.StorageClassAnalysis.DataExport.Destination != nil && b.StorageClassAnalysis.DataExport.Destination != nil {
				if ackcompare.HasNilDifference(a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination, b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination) {
					delta.Add("AnalyticsConfiguration.StorageClassAnalysis.DataExport.Destination.S3BucketDestination", a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination, b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination)
				} else if a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination != nil && b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination != nil {
					if ackcompare.HasNilDifference(a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Bucket, b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Bucket) {
						delta.Add("AnalyticsConfiguration.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Bucket", a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Bucket, b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Bucket)
					} else if a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Bucket != nil && b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Bucket != nil {
						if *a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Bucket != *b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Bucket {
							delta.Add("AnalyticsConfiguration.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Bucket", a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Bucket, b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Bucket)
						}
					}
					if ackcompare.HasNilDifference(a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.BucketAccountID, b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.BucketAccountID) {
						delta.Add("AnalyticsConfiguration.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.BucketAccountID", a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.BucketAccountID, b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.BucketAccountID)
					} else if a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.BucketAccountID != nil && b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.BucketAccountID != nil {
						if *a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.BucketAccountID != *b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.BucketAccountID {
							delta.Add("AnalyticsConfiguration.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.BucketAccountID", a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.BucketAccountID, b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.BucketAccountID)
						}
					}
					if ackcompare.HasNilDifference(a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Format, b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Format) {
						delta.Add("AnalyticsConfiguration.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Format", a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Format, b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Format)
					} else if a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Format != nil && b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Format != nil {
						if *a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Format != *b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Format {
							delta.Add("AnalyticsConfiguration.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Format", a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Format, b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Format)
						}
					}
					if ackcompare.HasNilDifference(a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Prefix, b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Prefix) {
						delta.Add("AnalyticsConfiguration.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Prefix", a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Prefix, b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Prefix)
					} else if a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Prefix != nil && b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Prefix != nil {
						if *a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Prefix != *b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Prefix {
							delta.Add("AnalyticsConfiguration.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Prefix", a.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Prefix, b.StorageClassAnalysis.DataExport.Destination.S3BucketDestination.Prefix)
						}
					}
				}
			}
			if ackcompare.HasNilDifference(a.StorageClassAnalysis.DataExport.OutputSchemaVersion, b.StorageClassAnalysis.DataExport.OutputSchemaVersion) {
				delta.Add("AnalyticsConfiguration.StorageClassAnalysis.DataExport.OutputSchemaVersion", a.StorageClassAnalysis.DataExport.OutputSchemaVersion, b.StorageClassAnalysis.DataExport.OutputSchemaVersion)
			} else if a.StorageClassAnalysis.DataExport.OutputSchemaVersion != nil && b.StorageClassAnalysis.DataExport.OutputSchemaVersion != nil {
				if *a.StorageClassAnalysis.DataExport.OutputSchemaVersion != *b.StorageClassAnalysis.DataExport.OutputSchemaVersion {
					delta.Add("AnalyticsConfiguration.StorageClassAnalysis.DataExport.OutputSchemaVersion", a.StorageClassAnalysis.DataExport.OutputSchemaVersion, b.StorageClassAnalysis.DataExport.OutputSchemaVersion)
				}
			}
		}
	}

	return delta
}

// getAnalyticsConfigurationAction returns the determined action for a given
// configuration object, depending on the desired and latest values
func getAnalyticsConfigurationAction(
	c *svcapitypes.AnalyticsConfiguration,
	latest *resource,
) ConfigurationAction {
	action := ConfigurationActionPut
	if latest != nil {
		for _, l := range latest.ko.Spec.Analytics {
			if *l.ID != *c.ID {
				continue
			}

			// Don't perform any action if they are identical
			delta := compareAnalyticsConfiguration(l, c)
			if len(delta.Differences) > 0 {
				action = ConfigurationActionUpdate
			} else {
				action = ConfigurationActionNone
			}
			break
		}
	}
	return action
}

func (rm *resourceManager) newListBucketAnalyticsPayload(
	r *resource,
) *svcsdk.ListBucketAnalyticsConfigurationsInput {
	res := &svcsdk.ListBucketAnalyticsConfigurationsInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) newPutBucketAnalyticsPayload(
	r *resource,
	c svcapitypes.AnalyticsConfiguration,
) *svcsdk.PutBucketAnalyticsConfigurationInput {
	res := &svcsdk.PutBucketAnalyticsConfigurationInput{}
	res.Bucket = r.ko.Spec.Name
	res.Id = c.ID
	res.AnalyticsConfiguration = rm.newAnalyticsConfiguration(&c)

	return res
}

func (rm *resourceManager) newDeleteBucketAnalyticsPayload(
	r *resource,
	c svcapitypes.AnalyticsConfiguration,
) *svcsdk.DeleteBucketAnalyticsConfigurationInput {
	res := &svcsdk.DeleteBucketAnalyticsConfigurationInput{}
	res.Bucket = r.ko.Spec.Name
	res.Id = c.ID

	return res
}

func (rm *resourceManager) deleteAnalyticsConfiguration(
	ctx context.Context,
	r *resource,
	c svcapitypes.AnalyticsConfiguration,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.deleteAnalyticsConfiguration")
	defer exit(err)

	input := rm.newDeleteBucketAnalyticsPayload(r, c)
	_, err = rm.sdkapi.DeleteBucketAnalyticsConfiguration(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeleteBucketAnalyticsConfiguration", err)
	return err
}

func (rm *resourceManager) putAnalyticsConfiguration(
	ctx context.Context,
	r *resource,
	c svcapitypes.AnalyticsConfiguration,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.putAnalyticsConfiguration")
	defer exit(err)

	input := rm.newPutBucketAnalyticsPayload(r, c)
	_, err = rm.sdkapi.PutBucketAnalyticsConfiguration(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketAnalyticsConfiguration", err)
	return err
}

func (rm *resourceManager) syncAnalytics(
	ctx context.Context,
	desired *resource,
	latest *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncAnalytics")
	defer exit(err)

	for _, c := range desired.ko.Spec.Analytics {
		action := getAnalyticsConfigurationAction(c, latest)

		switch action {
		case ConfigurationActionUpdate:
			fallthrough
		case ConfigurationActionPut:
			if err = rm.putAnalyticsConfiguration(ctx, desired, *c); err != nil {
				return err
			}
		default:
		}
	}

	if latest != nil {
		// Find any configurations that are in the latest but not in desired
		for _, l := range latest.ko.Spec.Analytics {
			exists := false
			for _, c := range desired.ko.Spec.Analytics {
				if *c.ID != *l.ID {
					continue
				}
				exists = true
				break
			}

			if !exists {
				if err = rm.deleteAnalyticsConfiguration(ctx, desired, *l); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// newCORSConfiguration returns a CORSConfiguration object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newCORSConfiguration(
	r *resource,
) *svcsdktypes.CORSConfiguration {
	res := &svcsdktypes.CORSConfiguration{}

	if r.ko.Spec.CORS.CORSRules != nil {
		resf0 := []svcsdktypes.CORSRule{}
		for _, resf0iter := range r.ko.Spec.CORS.CORSRules {
			resf0elem := &svcsdktypes.CORSRule{}
			if resf0iter.AllowedHeaders != nil {
				resf0elem.AllowedHeaders = aws.ToStringSlice(resf0iter.AllowedHeaders)
			}
			if resf0iter.AllowedMethods != nil {
				resf0elem.AllowedMethods = aws.ToStringSlice(resf0iter.AllowedMethods)
			}
			if resf0iter.AllowedOrigins != nil {
				resf0elem.AllowedOrigins = aws.ToStringSlice(resf0iter.AllowedOrigins)
			}
			if resf0iter.ExposeHeaders != nil {
				resf0elem.ExposeHeaders = aws.ToStringSlice(resf0iter.ExposeHeaders)
			}
			if resf0iter.ID != nil {
				resf0elem.ID = resf0iter.ID
			}
			if resf0iter.MaxAgeSeconds != nil {
				maxAgeSecondsCopy := int32(*resf0iter.MaxAgeSeconds)
				resf0elem.MaxAgeSeconds = &maxAgeSecondsCopy
			}
			resf0 = append(resf0, *resf0elem)
		}
		res.CORSRules = resf0
	}

	return res
}

// setResourceCORS sets the `CORS` spec field
// given the output of a `GetBucketCors` operation.
func (rm *resourceManager) setResourceCORS(
	r *resource,
	resp *svcsdk.GetBucketCorsOutput,
) *svcapitypes.CORSConfiguration {
	res := &svcapitypes.CORSConfiguration{}

	if resp.CORSRules != nil {
		resf0 := []*svcapitypes.CORSRule{}
		for _, resf0iter := range resp.CORSRules {
			resf0elem := &svcapitypes.CORSRule{}
			if resf0iter.AllowedHeaders != nil {
				resf0elem.AllowedHeaders = aws.StringSlice(resf0iter.AllowedHeaders)
			}
			if resf0iter.AllowedMethods != nil {
				resf0elem.AllowedMethods = aws.StringSlice(resf0iter.AllowedMethods)
			}
			if resf0iter.AllowedOrigins != nil {
				resf0elem.AllowedOrigins = aws.StringSlice(resf0iter.AllowedOrigins)
			}
			if resf0iter.ExposeHeaders != nil {
				resf0elem.ExposeHeaders = aws.StringSlice(resf0iter.ExposeHeaders)
			}
			if resf0iter.ID != nil {
				resf0elem.ID = resf0iter.ID
			}
			if resf0iter.MaxAgeSeconds != nil {
				maxAgeSecondsCopy := int64(*resf0iter.MaxAgeSeconds)
				resf0elem.MaxAgeSeconds = &maxAgeSecondsCopy
			}
			resf0 = append(resf0, resf0elem)
		}
		res.CORSRules = resf0
	}

	return res
}

// newServerSideEncryptionConfiguration returns a ServerSideEncryptionConfiguration object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newServerSideEncryptionConfiguration(
	r *resource,
) *svcsdktypes.ServerSideEncryptionConfiguration {
	res := &svcsdktypes.ServerSideEncryptionConfiguration{}

	if r.ko.Spec.Encryption.Rules != nil {
		resf0 := []svcsdktypes.ServerSideEncryptionRule{}
		for _, resf0iter := range r.ko.Spec.Encryption.Rules {
			resf0elem := &svcsdktypes.ServerSideEncryptionRule{}
			if resf0iter.ApplyServerSideEncryptionByDefault != nil {
				resf0elemf0 := &svcsdktypes.ServerSideEncryptionByDefault{}
				if resf0iter.ApplyServerSideEncryptionByDefault.KMSMasterKeyID != nil {
					resf0elemf0.KMSMasterKeyID = resf0iter.ApplyServerSideEncryptionByDefault.KMSMasterKeyID
				}
				if resf0iter.ApplyServerSideEncryptionByDefault.SSEAlgorithm != nil {
					resf0elemf0.SSEAlgorithm = svcsdktypes.ServerSideEncryption(*resf0iter.ApplyServerSideEncryptionByDefault.SSEAlgorithm)
				}
				resf0elem.ApplyServerSideEncryptionByDefault = resf0elemf0
			}
			if resf0iter.BucketKeyEnabled != nil {
				resf0elem.BucketKeyEnabled = resf0iter.BucketKeyEnabled
			}
			resf0 = append(resf0, *resf0elem)
		}
		res.Rules = resf0
	}

	return res
}

// setResourceEncryption sets the `Encryption` spec field
// given the output of a `GetBucketEncryption` operation.
func (rm *resourceManager) setResourceEncryption(
	r *resource,
	resp *svcsdk.GetBucketEncryptionOutput,
) *svcapitypes.ServerSideEncryptionConfiguration {
	res := &svcapitypes.ServerSideEncryptionConfiguration{}

	if resp.ServerSideEncryptionConfiguration.Rules != nil {
		resf0 := []*svcapitypes.ServerSideEncryptionRule{}
		for _, resf0iter := range resp.ServerSideEncryptionConfiguration.Rules {
			resf0elem := &svcapitypes.ServerSideEncryptionRule{}
			if resf0iter.ApplyServerSideEncryptionByDefault != nil {
				resf0elemf0 := &svcapitypes.ServerSideEncryptionByDefault{}
				if resf0iter.ApplyServerSideEncryptionByDefault.KMSMasterKeyID != nil {
					resf0elemf0.KMSMasterKeyID = resf0iter.ApplyServerSideEncryptionByDefault.KMSMasterKeyID
				}
				if resf0iter.ApplyServerSideEncryptionByDefault.SSEAlgorithm != "" {
					resf0elemf0.SSEAlgorithm = aws.String(string(resf0iter.ApplyServerSideEncryptionByDefault.SSEAlgorithm))
				}
				resf0elem.ApplyServerSideEncryptionByDefault = resf0elemf0
			}
			if resf0iter.BucketKeyEnabled != nil {
				resf0elem.BucketKeyEnabled = resf0iter.BucketKeyEnabled
			}
			resf0 = append(resf0, resf0elem)
		}
		res.Rules = resf0
	}

	return res
}

// newIntelligentTieringConfiguration returns a IntelligentTieringConfiguration object
// with each the field set by the corresponding configuration's fields.
func (rm *resourceManager) newIntelligentTieringConfiguration(
	c *svcapitypes.IntelligentTieringConfiguration,
) *svcsdktypes.IntelligentTieringConfiguration {
	res := &svcsdktypes.IntelligentTieringConfiguration{}

	if c.Filter != nil {
		resf0 := &svcsdktypes.IntelligentTieringFilter{}
		if c.Filter.And != nil {
			resf0f0 := &svcsdktypes.IntelligentTieringAndOperator{}
			if c.Filter.And.Prefix != nil {
				resf0f0.Prefix = c.Filter.And.Prefix
			}
			if c.Filter.And.Tags != nil {
				resf0f0f1 := []svcsdktypes.Tag{}
				for _, resf0f0f1iter := range c.Filter.And.Tags {
					resf0f0f1elem := &svcsdktypes.Tag{}
					if resf0f0f1iter.Key != nil {
						resf0f0f1elem.Key = resf0f0f1iter.Key
					}
					if resf0f0f1iter.Value != nil {
						resf0f0f1elem.Value = resf0f0f1iter.Value
					}
					resf0f0f1 = append(resf0f0f1, *resf0f0f1elem)
				}
				resf0f0.Tags = resf0f0f1
			}
			resf0.And = resf0f0
		}
		if c.Filter.Prefix != nil {
			resf0.Prefix = c.Filter.Prefix
		}
		if c.Filter.Tag != nil {
			resf0f2 := &svcsdktypes.Tag{}
			if c.Filter.Tag.Key != nil {
				resf0f2.Key = c.Filter.Tag.Key
			}
			if c.Filter.Tag.Value != nil {
				resf0f2.Value = c.Filter.Tag.Value
			}
			resf0.Tag = resf0f2
		}
		res.Filter = resf0
	}
	if c.ID != nil {
		res.Id = c.ID
	}
	if c.Status != nil {
		res.Status = svcsdktypes.IntelligentTieringStatus(*c.Status)
	}
	if c.Tierings != nil {
		resf3 := []svcsdktypes.Tiering{}
		for _, resf3iter := range c.Tierings {
			resf3elem := &svcsdktypes.Tiering{}
			if resf3iter.AccessTier != nil {
				resf3elem.AccessTier = svcsdktypes.IntelligentTieringAccessTier(*resf3iter.AccessTier)
			}
			if resf3iter.Days != nil {
				daysCopy := int32(*resf3iter.Days)
				resf3elem.Days = &daysCopy
			}
			resf3 = append(resf3, *resf3elem)
		}
		res.Tierings = resf3
	}

	return res
}

// setIntelligentTieringConfiguration sets a resource IntelligentTieringConfiguration type
// given the SDK type.
func (rm *resourceManager) setResourceIntelligentTieringConfiguration(
	r *resource,
	resp svcsdktypes.IntelligentTieringConfiguration,
) *svcapitypes.IntelligentTieringConfiguration {
	res := &svcapitypes.IntelligentTieringConfiguration{}

	if resp.Filter != nil {
		resf0 := &svcapitypes.IntelligentTieringFilter{}
		if resp.Filter.And != nil {
			resf0f0 := &svcapitypes.IntelligentTieringAndOperator{}
			if resp.Filter.And.Prefix != nil {
				resf0f0.Prefix = resp.Filter.And.Prefix
			}
			if resp.Filter.And.Tags != nil {
				resf0f0f1 := []*svcapitypes.Tag{}
				for _, resf0f0f1iter := range resp.Filter.And.Tags {
					resf0f0f1elem := &svcapitypes.Tag{}
					if resf0f0f1iter.Key != nil {
						resf0f0f1elem.Key = resf0f0f1iter.Key
					}
					if resf0f0f1iter.Value != nil {
						resf0f0f1elem.Value = resf0f0f1iter.Value
					}
					resf0f0f1 = append(resf0f0f1, resf0f0f1elem)
				}
				resf0f0.Tags = resf0f0f1
			}
			resf0.And = resf0f0
		}
		if resp.Filter.Prefix != nil {
			resf0.Prefix = resp.Filter.Prefix
		}
		if resp.Filter.Tag != nil {
			resf0f2 := &svcapitypes.Tag{}
			if resp.Filter.Tag.Key != nil {
				resf0f2.Key = resp.Filter.Tag.Key
			}
			if resp.Filter.Tag.Value != nil {
				resf0f2.Value = resp.Filter.Tag.Value
			}
			resf0.Tag = resf0f2
		}
		res.Filter = resf0
	}
	if resp.Id != nil {
		res.ID = resp.Id
	}
	if resp.Status != "" {
		res.Status = aws.String(string(resp.Status))
	}
	if resp.Tierings != nil {
		resf3 := []*svcapitypes.Tiering{}
		for _, resf3iter := range resp.Tierings {
			resf3elem := &svcapitypes.Tiering{}
			if resf3iter.AccessTier != "" {
				resf3elem.AccessTier = aws.String(string(resf3iter.AccessTier))
			}
			if resf3iter.Days != nil {
				daysCopy := int64(*resf3iter.Days)
				resf3elem.Days = &daysCopy
			}
			resf3 = append(resf3, resf3elem)
		}
		res.Tierings = resf3
	}

	return res
}

func compareIntelligentTieringConfiguration(
	a *svcapitypes.IntelligentTieringConfiguration,
	b *svcapitypes.IntelligentTieringConfiguration,
) *ackcompare.Delta {
	delta := ackcompare.NewDelta()
	if ackcompare.HasNilDifference(a.Filter, b.Filter) {
		delta.Add("IntelligentTieringConfiguration.Filter", a.Filter, b.Filter)
	} else if a.Filter != nil && b.Filter != nil {
		if ackcompare.HasNilDifference(a.Filter.And, b.Filter.And) {
			delta.Add("IntelligentTieringConfiguration.Filter.And", a.Filter.And, b.Filter.And)
		} else if a.Filter.And != nil && b.Filter.And != nil {
			if ackcompare.HasNilDifference(a.Filter.And.Prefix, b.Filter.And.Prefix) {
				delta.Add("IntelligentTieringConfiguration.Filter.And.Prefix", a.Filter.And.Prefix, b.Filter.And.Prefix)
			} else if a.Filter.And.Prefix != nil && b.Filter.And.Prefix != nil {
				if *a.Filter.And.Prefix != *b.Filter.And.Prefix {
					delta.Add("IntelligentTieringConfiguration.Filter.And.Prefix", a.Filter.And.Prefix, b.Filter.And.Prefix)
				}
			}
			if len(a.Filter.And.Tags) != len(b.Filter.And.Tags) {
				delta.Add("IntelligentTieringConfiguration.Filter.And.Tags", a.Filter.And.Tags, b.Filter.And.Tags)
			} else if len(a.Filter.And.Tags) > 0 {
				if !reflect.DeepEqual(a.Filter.And.Tags, b.Filter.And.Tags) {
					delta.Add("IntelligentTieringConfiguration.Filter.And.Tags", a.Filter.And.Tags, b.Filter.And.Tags)
				}
			}
		}
		if ackcompare.HasNilDifference(a.Filter.Prefix, b.Filter.Prefix) {
			delta.Add("IntelligentTieringConfiguration.Filter.Prefix", a.Filter.Prefix, b.Filter.Prefix)
		} else if a.Filter.Prefix != nil && b.Filter.Prefix != nil {
			if *a.Filter.Prefix != *b.Filter.Prefix {
				delta.Add("IntelligentTieringConfiguration.Filter.Prefix", a.Filter.Prefix, b.Filter.Prefix)
			}
		}
		if ackcompare.HasNilDifference(a.Filter.Tag, b.Filter.Tag) {
			delta.Add("IntelligentTieringConfiguration.Filter.Tag", a.Filter.Tag, b.Filter.Tag)
		} else if a.Filter.Tag != nil && b.Filter.Tag != nil {
			if ackcompare.HasNilDifference(a.Filter.Tag.Key, b.Filter.Tag.Key) {
				delta.Add("IntelligentTieringConfiguration.Filter.Tag.Key", a.Filter.Tag.Key, b.Filter.Tag.Key)
			} else if a.Filter.Tag.Key != nil && b.Filter.Tag.Key != nil {
				if *a.Filter.Tag.Key != *b.Filter.Tag.Key {
					delta.Add("IntelligentTieringConfiguration.Filter.Tag.Key", a.Filter.Tag.Key, b.Filter.Tag.Key)
				}
			}
			if ackcompare.HasNilDifference(a.Filter.Tag.Value, b.Filter.Tag.Value) {
				delta.Add("IntelligentTieringConfiguration.Filter.Tag.Value", a.Filter.Tag.Value, b.Filter.Tag.Value)
			} else if a.Filter.Tag.Value != nil && b.Filter.Tag.Value != nil {
				if *a.Filter.Tag.Value != *b.Filter.Tag.Value {
					delta.Add("IntelligentTieringConfiguration.Filter.Tag.Value", a.Filter.Tag.Value, b.Filter.Tag.Value)
				}
			}
		}
	}
	if ackcompare.HasNilDifference(a.ID, b.ID) {
		delta.Add("IntelligentTieringConfiguration.ID", a.ID, b.ID)
	} else if a.ID != nil && b.ID != nil {
		if *a.ID != *b.ID {
			delta.Add("IntelligentTieringConfiguration.ID", a.ID, b.ID)
		}
	}
	if ackcompare.HasNilDifference(a.Status, b.Status) {
		delta.Add("IntelligentTieringConfiguration.Status", a.Status, b.Status)
	} else if a.Status != nil && b.Status != nil {
		if *a.Status != *b.Status {
			delta.Add("IntelligentTieringConfiguration.Status", a.Status, b.Status)
		}
	}
	if len(a.Tierings) != len(b.Tierings) {
		delta.Add("IntelligentTieringConfiguration.Tierings", a.Tierings, b.Tierings)
	} else if len(a.Tierings) > 0 {
		if !reflect.DeepEqual(a.Tierings, b.Tierings) {
			delta.Add("IntelligentTieringConfiguration.Tierings", a.Tierings, b.Tierings)
		}
	}

	return delta
}

// getIntelligentTieringConfigurationAction returns the determined action for a given
// configuration object, depending on the desired and latest values
func getIntelligentTieringConfigurationAction(
	c *svcapitypes.IntelligentTieringConfiguration,
	latest *resource,
) ConfigurationAction {
	action := ConfigurationActionPut
	if latest != nil {
		for _, l := range latest.ko.Spec.IntelligentTiering {
			if *l.ID != *c.ID {
				continue
			}

			// Don't perform any action if they are identical
			delta := compareIntelligentTieringConfiguration(l, c)
			if len(delta.Differences) > 0 {
				action = ConfigurationActionUpdate
			} else {
				action = ConfigurationActionNone
			}
			break
		}
	}
	return action
}

func (rm *resourceManager) newListBucketIntelligentTieringPayload(
	r *resource,
) *svcsdk.ListBucketIntelligentTieringConfigurationsInput {
	res := &svcsdk.ListBucketIntelligentTieringConfigurationsInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) newPutBucketIntelligentTieringPayload(
	r *resource,
	c svcapitypes.IntelligentTieringConfiguration,
) *svcsdk.PutBucketIntelligentTieringConfigurationInput {
	res := &svcsdk.PutBucketIntelligentTieringConfigurationInput{}
	res.Bucket = r.ko.Spec.Name
	res.Id = c.ID
	res.IntelligentTieringConfiguration = rm.newIntelligentTieringConfiguration(&c)

	return res
}

func (rm *resourceManager) newDeleteBucketIntelligentTieringPayload(
	r *resource,
	c svcapitypes.IntelligentTieringConfiguration,
) *svcsdk.DeleteBucketIntelligentTieringConfigurationInput {
	res := &svcsdk.DeleteBucketIntelligentTieringConfigurationInput{}
	res.Bucket = r.ko.Spec.Name
	res.Id = c.ID

	return res
}

func (rm *resourceManager) deleteIntelligentTieringConfiguration(
	ctx context.Context,
	r *resource,
	c svcapitypes.IntelligentTieringConfiguration,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.deleteIntelligentTieringConfiguration")
	defer exit(err)

	input := rm.newDeleteBucketIntelligentTieringPayload(r, c)
	_, err = rm.sdkapi.DeleteBucketIntelligentTieringConfiguration(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeleteBucketIntelligentTieringConfiguration", err)
	return err
}

func (rm *resourceManager) putIntelligentTieringConfiguration(
	ctx context.Context,
	r *resource,
	c svcapitypes.IntelligentTieringConfiguration,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.putIntelligentTieringConfiguration")
	defer exit(err)

	input := rm.newPutBucketIntelligentTieringPayload(r, c)
	_, err = rm.sdkapi.PutBucketIntelligentTieringConfiguration(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketIntelligentTieringConfiguration", err)
	return err
}

func (rm *resourceManager) syncIntelligentTiering(
	ctx context.Context,
	desired *resource,
	latest *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncIntelligentTiering")
	defer exit(err)

	for _, c := range desired.ko.Spec.IntelligentTiering {
		action := getIntelligentTieringConfigurationAction(c, latest)

		switch action {
		case ConfigurationActionUpdate:
			fallthrough
		case ConfigurationActionPut:
			if err = rm.putIntelligentTieringConfiguration(ctx, desired, *c); err != nil {
				return err
			}
		default:
		}
	}

	if latest != nil {
		// Find any configurations that are in the latest but not in desired
		for _, l := range latest.ko.Spec.IntelligentTiering {
			exists := false
			for _, c := range desired.ko.Spec.IntelligentTiering {
				if *c.ID != *l.ID {
					continue
				}
				exists = true
				break
			}

			if !exists {
				if err = rm.deleteIntelligentTieringConfiguration(ctx, desired, *l); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// newInventoryConfiguration returns a InventoryConfiguration object
// with each the field set by the corresponding configuration's fields.
func (rm *resourceManager) newInventoryConfiguration(
	c *svcapitypes.InventoryConfiguration,
) *svcsdktypes.InventoryConfiguration {
	res := &svcsdktypes.InventoryConfiguration{}

	if c.Destination != nil {
		resf0 := &svcsdktypes.InventoryDestination{}
		if c.Destination.S3BucketDestination != nil {
			resf0f0 := &svcsdktypes.InventoryS3BucketDestination{}
			if c.Destination.S3BucketDestination.AccountID != nil {
				resf0f0.AccountId = c.Destination.S3BucketDestination.AccountID
			}
			if c.Destination.S3BucketDestination.Bucket != nil {
				resf0f0.Bucket = c.Destination.S3BucketDestination.Bucket
			}
			if c.Destination.S3BucketDestination.Encryption != nil {
				resf0f0f2 := &svcsdktypes.InventoryEncryption{}
				if c.Destination.S3BucketDestination.Encryption.SSEKMS != nil {
					resf0f0f2f0 := &svcsdktypes.SSEKMS{}
					if c.Destination.S3BucketDestination.Encryption.SSEKMS.KeyID != nil {
						resf0f0f2f0.KeyId = c.Destination.S3BucketDestination.Encryption.SSEKMS.KeyID
					}
					resf0f0f2.SSEKMS = resf0f0f2f0
				}
				resf0f0.Encryption = resf0f0f2
			}
			if c.Destination.S3BucketDestination.Format != nil {
				resf0f0.Format = svcsdktypes.InventoryFormat(*c.Destination.S3BucketDestination.Format)
			}
			if c.Destination.S3BucketDestination.Prefix != nil {
				resf0f0.Prefix = c.Destination.S3BucketDestination.Prefix
			}
			resf0.S3BucketDestination = resf0f0
		}
		res.Destination = resf0
	}
	if c.Filter != nil {
		resf1 := &svcsdktypes.InventoryFilter{}
		if c.Filter.Prefix != nil {
			resf1.Prefix = c.Filter.Prefix
		}
		res.Filter = resf1
	}
	if c.ID != nil {
		res.Id = c.ID
	}
	if c.IncludedObjectVersions != nil {
		res.IncludedObjectVersions = svcsdktypes.InventoryIncludedObjectVersions(*c.IncludedObjectVersions)
	}
	if c.IsEnabled != nil {
		res.IsEnabled = c.IsEnabled
	}
	if c.OptionalFields != nil {
		resf5 := []svcsdktypes.InventoryOptionalField{}
		for _, resf5iter := range c.OptionalFields {
			var resf5elem string
			resf5elem = string(*resf5iter)
			resf5 = append(resf5, svcsdktypes.InventoryOptionalField(resf5elem))
		}
		res.OptionalFields = resf5
	}
	if c.Schedule != nil {
		resf6 := &svcsdktypes.InventorySchedule{}
		if c.Schedule.Frequency != nil {
			resf6.Frequency = svcsdktypes.InventoryFrequency(*c.Schedule.Frequency)
		}
		res.Schedule = resf6
	}

	return res
}

// setInventoryConfiguration sets a resource InventoryConfiguration type
// given the SDK type.
func (rm *resourceManager) setResourceInventoryConfiguration(
	r *resource,
	resp svcsdktypes.InventoryConfiguration,
) *svcapitypes.InventoryConfiguration {
	res := &svcapitypes.InventoryConfiguration{}

	if resp.Destination != nil {
		resf0 := &svcapitypes.InventoryDestination{}
		if resp.Destination.S3BucketDestination != nil {
			resf0f0 := &svcapitypes.InventoryS3BucketDestination{}
			if resp.Destination.S3BucketDestination.AccountId != nil {
				resf0f0.AccountID = resp.Destination.S3BucketDestination.AccountId
			}
			if resp.Destination.S3BucketDestination.Bucket != nil {
				resf0f0.Bucket = resp.Destination.S3BucketDestination.Bucket
			}
			if resp.Destination.S3BucketDestination.Encryption != nil {
				resf0f0f2 := &svcapitypes.InventoryEncryption{}
				if resp.Destination.S3BucketDestination.Encryption.SSEKMS != nil {
					resf0f0f2f0 := &svcapitypes.SSEKMS{}
					if resp.Destination.S3BucketDestination.Encryption.SSEKMS.KeyId != nil {
						resf0f0f2f0.KeyID = resp.Destination.S3BucketDestination.Encryption.SSEKMS.KeyId
					}
					resf0f0f2.SSEKMS = resf0f0f2f0
				}
				resf0f0.Encryption = resf0f0f2
			}
			if resp.Destination.S3BucketDestination.Format != "" {
				resf0f0.Format = aws.String(string(resp.Destination.S3BucketDestination.Format))
			}
			if resp.Destination.S3BucketDestination.Prefix != nil {
				resf0f0.Prefix = resp.Destination.S3BucketDestination.Prefix
			}
			resf0.S3BucketDestination = resf0f0
		}
		res.Destination = resf0
	}
	if resp.Filter != nil {
		resf1 := &svcapitypes.InventoryFilter{}
		if resp.Filter.Prefix != nil {
			resf1.Prefix = resp.Filter.Prefix
		}
		res.Filter = resf1
	}
	if resp.Id != nil {
		res.ID = resp.Id
	}
	if resp.IncludedObjectVersions != "" {
		res.IncludedObjectVersions = aws.String(string(resp.IncludedObjectVersions))
	}
	if resp.IsEnabled != nil {
		res.IsEnabled = resp.IsEnabled
	}
	if resp.OptionalFields != nil {
		resf5 := []*string{}
		for _, resf5iter := range resp.OptionalFields {
			var resf5elem *string
			resf5elem = aws.String(string(resf5iter))
			resf5 = append(resf5, resf5elem)
		}
		res.OptionalFields = resf5
	}
	if resp.Schedule != nil {
		resf6 := &svcapitypes.InventorySchedule{}
		if resp.Schedule.Frequency != "" {
			resf6.Frequency = aws.String(string(resp.Schedule.Frequency))
		}
		res.Schedule = resf6
	}

	return res
}

func compareInventoryConfiguration(
	a *svcapitypes.InventoryConfiguration,
	b *svcapitypes.InventoryConfiguration,
) *ackcompare.Delta {
	delta := ackcompare.NewDelta()
	if ackcompare.HasNilDifference(a.Destination, b.Destination) {
		delta.Add("InventoryConfiguration.Destination", a.Destination, b.Destination)
	} else if a.Destination != nil && b.Destination != nil {
		if ackcompare.HasNilDifference(a.Destination.S3BucketDestination, b.Destination.S3BucketDestination) {
			delta.Add("InventoryConfiguration.Destination.S3BucketDestination", a.Destination.S3BucketDestination, b.Destination.S3BucketDestination)
		} else if a.Destination.S3BucketDestination != nil && b.Destination.S3BucketDestination != nil {
			if ackcompare.HasNilDifference(a.Destination.S3BucketDestination.AccountID, b.Destination.S3BucketDestination.AccountID) {
				delta.Add("InventoryConfiguration.Destination.S3BucketDestination.AccountID", a.Destination.S3BucketDestination.AccountID, b.Destination.S3BucketDestination.AccountID)
			} else if a.Destination.S3BucketDestination.AccountID != nil && b.Destination.S3BucketDestination.AccountID != nil {
				if *a.Destination.S3BucketDestination.AccountID != *b.Destination.S3BucketDestination.AccountID {
					delta.Add("InventoryConfiguration.Destination.S3BucketDestination.AccountID", a.Destination.S3BucketDestination.AccountID, b.Destination.S3BucketDestination.AccountID)
				}
			}
			if ackcompare.HasNilDifference(a.Destination.S3BucketDestination.Bucket, b.Destination.S3BucketDestination.Bucket) {
				delta.Add("InventoryConfiguration.Destination.S3BucketDestination.Bucket", a.Destination.S3BucketDestination.Bucket, b.Destination.S3BucketDestination.Bucket)
			} else if a.Destination.S3BucketDestination.Bucket != nil && b.Destination.S3BucketDestination.Bucket != nil {
				if *a.Destination.S3BucketDestination.Bucket != *b.Destination.S3BucketDestination.Bucket {
					delta.Add("InventoryConfiguration.Destination.S3BucketDestination.Bucket", a.Destination.S3BucketDestination.Bucket, b.Destination.S3BucketDestination.Bucket)
				}
			}
			if ackcompare.HasNilDifference(a.Destination.S3BucketDestination.Encryption, b.Destination.S3BucketDestination.Encryption) {
				delta.Add("InventoryConfiguration.Destination.S3BucketDestination.Encryption", a.Destination.S3BucketDestination.Encryption, b.Destination.S3BucketDestination.Encryption)
			} else if a.Destination.S3BucketDestination.Encryption != nil && b.Destination.S3BucketDestination.Encryption != nil {
				if ackcompare.HasNilDifference(a.Destination.S3BucketDestination.Encryption.SSEKMS, b.Destination.S3BucketDestination.Encryption.SSEKMS) {
					delta.Add("InventoryConfiguration.Destination.S3BucketDestination.Encryption.SSEKMS", a.Destination.S3BucketDestination.Encryption.SSEKMS, b.Destination.S3BucketDestination.Encryption.SSEKMS)
				} else if a.Destination.S3BucketDestination.Encryption.SSEKMS != nil && b.Destination.S3BucketDestination.Encryption.SSEKMS != nil {
					if ackcompare.HasNilDifference(a.Destination.S3BucketDestination.Encryption.SSEKMS.KeyID, b.Destination.S3BucketDestination.Encryption.SSEKMS.KeyID) {
						delta.Add("InventoryConfiguration.Destination.S3BucketDestination.Encryption.SSEKMS.KeyID", a.Destination.S3BucketDestination.Encryption.SSEKMS.KeyID, b.Destination.S3BucketDestination.Encryption.SSEKMS.KeyID)
					} else if a.Destination.S3BucketDestination.Encryption.SSEKMS.KeyID != nil && b.Destination.S3BucketDestination.Encryption.SSEKMS.KeyID != nil {
						if *a.Destination.S3BucketDestination.Encryption.SSEKMS.KeyID != *b.Destination.S3BucketDestination.Encryption.SSEKMS.KeyID {
							delta.Add("InventoryConfiguration.Destination.S3BucketDestination.Encryption.SSEKMS.KeyID", a.Destination.S3BucketDestination.Encryption.SSEKMS.KeyID, b.Destination.S3BucketDestination.Encryption.SSEKMS.KeyID)
						}
					}
				}
			}
			if ackcompare.HasNilDifference(a.Destination.S3BucketDestination.Format, b.Destination.S3BucketDestination.Format) {
				delta.Add("InventoryConfiguration.Destination.S3BucketDestination.Format", a.Destination.S3BucketDestination.Format, b.Destination.S3BucketDestination.Format)
			} else if a.Destination.S3BucketDestination.Format != nil && b.Destination.S3BucketDestination.Format != nil {
				if *a.Destination.S3BucketDestination.Format != *b.Destination.S3BucketDestination.Format {
					delta.Add("InventoryConfiguration.Destination.S3BucketDestination.Format", a.Destination.S3BucketDestination.Format, b.Destination.S3BucketDestination.Format)
				}
			}
			if ackcompare.HasNilDifference(a.Destination.S3BucketDestination.Prefix, b.Destination.S3BucketDestination.Prefix) {
				delta.Add("InventoryConfiguration.Destination.S3BucketDestination.Prefix", a.Destination.S3BucketDestination.Prefix, b.Destination.S3BucketDestination.Prefix)
			} else if a.Destination.S3BucketDestination.Prefix != nil && b.Destination.S3BucketDestination.Prefix != nil {
				if *a.Destination.S3BucketDestination.Prefix != *b.Destination.S3BucketDestination.Prefix {
					delta.Add("InventoryConfiguration.Destination.S3BucketDestination.Prefix", a.Destination.S3BucketDestination.Prefix, b.Destination.S3BucketDestination.Prefix)
				}
			}
		}
	}
	if ackcompare.HasNilDifference(a.Filter, b.Filter) {
		delta.Add("InventoryConfiguration.Filter", a.Filter, b.Filter)
	} else if a.Filter != nil && b.Filter != nil {
		if ackcompare.HasNilDifference(a.Filter.Prefix, b.Filter.Prefix) {
			delta.Add("InventoryConfiguration.Filter.Prefix", a.Filter.Prefix, b.Filter.Prefix)
		} else if a.Filter.Prefix != nil && b.Filter.Prefix != nil {
			if *a.Filter.Prefix != *b.Filter.Prefix {
				delta.Add("InventoryConfiguration.Filter.Prefix", a.Filter.Prefix, b.Filter.Prefix)
			}
		}
	}
	if ackcompare.HasNilDifference(a.ID, b.ID) {
		delta.Add("InventoryConfiguration.ID", a.ID, b.ID)
	} else if a.ID != nil && b.ID != nil {
		if *a.ID != *b.ID {
			delta.Add("InventoryConfiguration.ID", a.ID, b.ID)
		}
	}
	if ackcompare.HasNilDifference(a.IncludedObjectVersions, b.IncludedObjectVersions) {
		delta.Add("InventoryConfiguration.IncludedObjectVersions", a.IncludedObjectVersions, b.IncludedObjectVersions)
	} else if a.IncludedObjectVersions != nil && b.IncludedObjectVersions != nil {
		if *a.IncludedObjectVersions != *b.IncludedObjectVersions {
			delta.Add("InventoryConfiguration.IncludedObjectVersions", a.IncludedObjectVersions, b.IncludedObjectVersions)
		}
	}
	if ackcompare.HasNilDifference(a.IsEnabled, b.IsEnabled) {
		delta.Add("InventoryConfiguration.IsEnabled", a.IsEnabled, b.IsEnabled)
	} else if a.IsEnabled != nil && b.IsEnabled != nil {
		if *a.IsEnabled != *b.IsEnabled {
			delta.Add("InventoryConfiguration.IsEnabled", a.IsEnabled, b.IsEnabled)
		}
	}
	if len(a.OptionalFields) != len(b.OptionalFields) {
		delta.Add("InventoryConfiguration.OptionalFields", a.OptionalFields, b.OptionalFields)
	} else if len(a.OptionalFields) > 0 {
		if !ackcompare.SliceStringPEqual(a.OptionalFields, b.OptionalFields) {
			delta.Add("InventoryConfiguration.OptionalFields", a.OptionalFields, b.OptionalFields)
		}
	}
	if ackcompare.HasNilDifference(a.Schedule, b.Schedule) {
		delta.Add("InventoryConfiguration.Schedule", a.Schedule, b.Schedule)
	} else if a.Schedule != nil && b.Schedule != nil {
		if ackcompare.HasNilDifference(a.Schedule.Frequency, b.Schedule.Frequency) {
			delta.Add("InventoryConfiguration.Schedule.Frequency", a.Schedule.Frequency, b.Schedule.Frequency)
		} else if a.Schedule.Frequency != nil && b.Schedule.Frequency != nil {
			if *a.Schedule.Frequency != *b.Schedule.Frequency {
				delta.Add("InventoryConfiguration.Schedule.Frequency", a.Schedule.Frequency, b.Schedule.Frequency)
			}
		}
	}

	return delta
}

// getInventoryConfigurationAction returns the determined action for a given
// configuration object, depending on the desired and latest values
func getInventoryConfigurationAction(
	c *svcapitypes.InventoryConfiguration,
	latest *resource,
) ConfigurationAction {
	action := ConfigurationActionPut
	if latest != nil {
		for _, l := range latest.ko.Spec.Inventory {
			if *l.ID != *c.ID {
				continue
			}

			// Don't perform any action if they are identical
			delta := compareInventoryConfiguration(l, c)
			if len(delta.Differences) > 0 {
				action = ConfigurationActionUpdate
			} else {
				action = ConfigurationActionNone
			}
			break
		}
	}
	return action
}

func (rm *resourceManager) newListBucketInventoryPayload(
	r *resource,
) *svcsdk.ListBucketInventoryConfigurationsInput {
	res := &svcsdk.ListBucketInventoryConfigurationsInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) newPutBucketInventoryPayload(
	r *resource,
	c svcapitypes.InventoryConfiguration,
) *svcsdk.PutBucketInventoryConfigurationInput {
	res := &svcsdk.PutBucketInventoryConfigurationInput{}
	res.Bucket = r.ko.Spec.Name
	res.Id = c.ID
	res.InventoryConfiguration = rm.newInventoryConfiguration(&c)

	return res
}

func (rm *resourceManager) newDeleteBucketInventoryPayload(
	r *resource,
	c svcapitypes.InventoryConfiguration,
) *svcsdk.DeleteBucketInventoryConfigurationInput {
	res := &svcsdk.DeleteBucketInventoryConfigurationInput{}
	res.Bucket = r.ko.Spec.Name
	res.Id = c.ID

	return res
}

func (rm *resourceManager) deleteInventoryConfiguration(
	ctx context.Context,
	r *resource,
	c svcapitypes.InventoryConfiguration,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.deleteInventoryConfiguration")
	defer exit(err)

	input := rm.newDeleteBucketInventoryPayload(r, c)
	_, err = rm.sdkapi.DeleteBucketInventoryConfiguration(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeleteBucketInventoryConfiguration", err)
	return err
}

func (rm *resourceManager) putInventoryConfiguration(
	ctx context.Context,
	r *resource,
	c svcapitypes.InventoryConfiguration,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.putInventoryConfiguration")
	defer exit(err)

	input := rm.newPutBucketInventoryPayload(r, c)
	_, err = rm.sdkapi.PutBucketInventoryConfiguration(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketInventoryConfiguration", err)
	return err
}

func (rm *resourceManager) syncInventory(
	ctx context.Context,
	desired *resource,
	latest *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncInventory")
	defer exit(err)

	for _, c := range desired.ko.Spec.Inventory {
		action := getInventoryConfigurationAction(c, latest)

		switch action {
		case ConfigurationActionUpdate:
			fallthrough
		case ConfigurationActionPut:
			if err = rm.putInventoryConfiguration(ctx, desired, *c); err != nil {
				return err
			}
		default:
		}
	}

	if latest != nil {
		// Find any configurations that are in the latest but not in desired
		for _, l := range latest.ko.Spec.Inventory {
			exists := false
			for _, c := range desired.ko.Spec.Inventory {
				if *c.ID != *l.ID {
					continue
				}
				exists = true
				break
			}

			if !exists {
				if err = rm.deleteInventoryConfiguration(ctx, desired, *l); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// newLifecycleConfiguration returns a LifecycleConfiguration object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newLifecycleConfiguration(
	r *resource,
) *svcsdktypes.BucketLifecycleConfiguration {
	res := &svcsdktypes.BucketLifecycleConfiguration{}

	if r.ko.Spec.Lifecycle.Rules != nil {
		resf0 := []svcsdktypes.LifecycleRule{}
		for _, resf0iter := range r.ko.Spec.Lifecycle.Rules {
			resf0elem := &svcsdktypes.LifecycleRule{}
			if resf0iter.AbortIncompleteMultipartUpload != nil {
				resf0elemf0 := &svcsdktypes.AbortIncompleteMultipartUpload{}
				if resf0iter.AbortIncompleteMultipartUpload.DaysAfterInitiation != nil {
					daysAfterInitiationCopy := int32(*resf0iter.AbortIncompleteMultipartUpload.DaysAfterInitiation)
					resf0elemf0.DaysAfterInitiation = &daysAfterInitiationCopy
				}
				resf0elem.AbortIncompleteMultipartUpload = resf0elemf0
			}
			if resf0iter.Expiration != nil {
				resf0elemf1 := &svcsdktypes.LifecycleExpiration{}
				if resf0iter.Expiration.Date != nil {
					resf0elemf1.Date = &resf0iter.Expiration.Date.Time
				}
				if resf0iter.Expiration.Days != nil {
					daysCopy := int32(*resf0iter.Expiration.Days)
					resf0elemf1.Days = &daysCopy
				}
				if resf0iter.Expiration.ExpiredObjectDeleteMarker != nil {
					resf0elemf1.ExpiredObjectDeleteMarker = resf0iter.Expiration.ExpiredObjectDeleteMarker
				}
				resf0elem.Expiration = resf0elemf1
			}
			if resf0iter.Filter != nil {
				resf0elemf2 := &svcsdktypes.LifecycleRuleFilter{}
				if resf0iter.Filter.And != nil {
					resf0elemf2f0 := &svcsdktypes.LifecycleRuleAndOperator{}
					if resf0iter.Filter.And.ObjectSizeGreaterThan != nil {
						resf0elemf2f0.ObjectSizeGreaterThan = resf0iter.Filter.And.ObjectSizeGreaterThan
					}
					if resf0iter.Filter.And.ObjectSizeLessThan != nil {
						resf0elemf2f0.ObjectSizeLessThan = resf0iter.Filter.And.ObjectSizeLessThan
					}
					if resf0iter.Filter.And.Prefix != nil {
						resf0elemf2f0.Prefix = resf0iter.Filter.And.Prefix
					}
					if resf0iter.Filter.And.Tags != nil {
						resf0elemf2f0f3 := []svcsdktypes.Tag{}
						for _, resf0elemf2f0f3iter := range resf0iter.Filter.And.Tags {
							resf0elemf2f0f3elem := &svcsdktypes.Tag{}
							if resf0elemf2f0f3iter.Key != nil {
								resf0elemf2f0f3elem.Key = resf0elemf2f0f3iter.Key
							}
							if resf0elemf2f0f3iter.Value != nil {
								resf0elemf2f0f3elem.Value = resf0elemf2f0f3iter.Value
							}
							resf0elemf2f0f3 = append(resf0elemf2f0f3, *resf0elemf2f0f3elem)
						}
						resf0elemf2f0.Tags = resf0elemf2f0f3
					}
					resf0elemf2.And = resf0elemf2f0
				}
				if resf0iter.Filter.ObjectSizeGreaterThan != nil {
					resf0elemf2.ObjectSizeGreaterThan = resf0iter.Filter.ObjectSizeGreaterThan
				}
				if resf0iter.Filter.ObjectSizeLessThan != nil {
					resf0elemf2.ObjectSizeLessThan = resf0iter.Filter.ObjectSizeLessThan
				}
				if resf0iter.Filter.Prefix != nil {
					resf0elemf2.Prefix = resf0iter.Filter.Prefix
				}
				if resf0iter.Filter.Tag != nil {
					resf0elemf2f4 := &svcsdktypes.Tag{}
					if resf0iter.Filter.Tag.Key != nil {
						resf0elemf2f4.Key = resf0iter.Filter.Tag.Key
					}
					if resf0iter.Filter.Tag.Value != nil {
						resf0elemf2f4.Value = resf0iter.Filter.Tag.Value
					}
					resf0elemf2.Tag = resf0elemf2f4
				}
				resf0elem.Filter = resf0elemf2
			}
			if resf0iter.ID != nil {
				resf0elem.ID = resf0iter.ID
			}
			if resf0iter.NoncurrentVersionExpiration != nil {
				resf0elemf4 := &svcsdktypes.NoncurrentVersionExpiration{}
				if resf0iter.NoncurrentVersionExpiration.NewerNoncurrentVersions != nil {
					newerNoncurrentVersionsCopy := int32(*resf0iter.NoncurrentVersionExpiration.NewerNoncurrentVersions)
					resf0elemf4.NewerNoncurrentVersions = &newerNoncurrentVersionsCopy
				}
				if resf0iter.NoncurrentVersionExpiration.NoncurrentDays != nil {
					noncurrentDaysCopy := int32(*resf0iter.NoncurrentVersionExpiration.NoncurrentDays)
					resf0elemf4.NoncurrentDays = &noncurrentDaysCopy
				}
				resf0elem.NoncurrentVersionExpiration = resf0elemf4
			}
			if resf0iter.NoncurrentVersionTransitions != nil {
				resf0elemf5 := []svcsdktypes.NoncurrentVersionTransition{}
				for _, resf0elemf5iter := range resf0iter.NoncurrentVersionTransitions {
					resf0elemf5elem := &svcsdktypes.NoncurrentVersionTransition{}
					if resf0elemf5iter.NewerNoncurrentVersions != nil {
						newerNoncurrentVersionsCopy := int32(*resf0elemf5iter.NewerNoncurrentVersions)
						resf0elemf5elem.NewerNoncurrentVersions = &newerNoncurrentVersionsCopy
					}
					if resf0elemf5iter.NoncurrentDays != nil {
						noncurrentDaysCopy := int32(*resf0elemf5iter.NoncurrentDays)
						resf0elemf5elem.NoncurrentDays = &noncurrentDaysCopy
					}
					if resf0elemf5iter.StorageClass != nil {
						resf0elemf5elem.StorageClass = svcsdktypes.TransitionStorageClass(*resf0elemf5iter.StorageClass)
					}
					resf0elemf5 = append(resf0elemf5, *resf0elemf5elem)
				}
				resf0elem.NoncurrentVersionTransitions = resf0elemf5
			}
			if resf0iter.Prefix != nil {
				resf0elem.Prefix = resf0iter.Prefix
			}
			if resf0iter.Status != nil {
				resf0elem.Status = svcsdktypes.ExpirationStatus(*resf0iter.Status)
			}
			if resf0iter.Transitions != nil {
				resf0elemf8 := []svcsdktypes.Transition{}
				for _, resf0elemf8iter := range resf0iter.Transitions {
					resf0elemf8elem := &svcsdktypes.Transition{}
					if resf0elemf8iter.Date != nil {
						resf0elemf8elem.Date = &resf0elemf8iter.Date.Time
					}
					if resf0elemf8iter.Days != nil {
						daysCopy := int32(*resf0elemf8iter.Days)
						resf0elemf8elem.Days = &daysCopy
					}
					if resf0elemf8iter.StorageClass != nil {
						resf0elemf8elem.StorageClass = svcsdktypes.TransitionStorageClass(*resf0elemf8iter.StorageClass)
					}
					resf0elemf8 = append(resf0elemf8, *resf0elemf8elem)
				}
				resf0elem.Transitions = resf0elemf8
			}
			resf0 = append(resf0, *resf0elem)
		}
		res.Rules = resf0
	}

	return res
}

// setResourceLifecycle sets the `Lifecycle` spec field
// given the output of a `GetBucketLifecycleConfiguration` operation.
func (rm *resourceManager) setResourceLifecycle(
	r *resource,
	resp *svcsdk.GetBucketLifecycleConfigurationOutput,
) *svcapitypes.BucketLifecycleConfiguration {
	res := &svcapitypes.BucketLifecycleConfiguration{}

	if resp.Rules != nil {
		resf0 := []*svcapitypes.LifecycleRule{}
		for _, resf0iter := range resp.Rules {
			resf0elem := &svcapitypes.LifecycleRule{}
			if resf0iter.AbortIncompleteMultipartUpload != nil {
				resf0elemf0 := &svcapitypes.AbortIncompleteMultipartUpload{}
				if resf0iter.AbortIncompleteMultipartUpload.DaysAfterInitiation != nil {
					daysAfterInitiationCopy := int64(*resf0iter.AbortIncompleteMultipartUpload.DaysAfterInitiation)
					resf0elemf0.DaysAfterInitiation = &daysAfterInitiationCopy
				}
				resf0elem.AbortIncompleteMultipartUpload = resf0elemf0
			}
			if resf0iter.Expiration != nil {
				resf0elemf1 := &svcapitypes.LifecycleExpiration{}
				if resf0iter.Expiration.Date != nil {
					resf0elemf1.Date = &metav1.Time{*resf0iter.Expiration.Date}
				}
				if resf0iter.Expiration.Days != nil {
					daysCopy := int64(*resf0iter.Expiration.Days)
					resf0elemf1.Days = &daysCopy
				}
				if resf0iter.Expiration.ExpiredObjectDeleteMarker != nil {
					resf0elemf1.ExpiredObjectDeleteMarker = resf0iter.Expiration.ExpiredObjectDeleteMarker
				}
				resf0elem.Expiration = resf0elemf1
			}
			if resf0iter.Filter != nil {
				resf0elemf2 := &svcapitypes.LifecycleRuleFilter{}
				if resf0iter.Filter.And != nil {
					resf0elemf2f0 := &svcapitypes.LifecycleRuleAndOperator{}
					if resf0iter.Filter.And.ObjectSizeGreaterThan != nil {
						resf0elemf2f0.ObjectSizeGreaterThan = resf0iter.Filter.And.ObjectSizeGreaterThan
					}
					if resf0iter.Filter.And.ObjectSizeLessThan != nil {
						resf0elemf2f0.ObjectSizeLessThan = resf0iter.Filter.And.ObjectSizeLessThan
					}
					if resf0iter.Filter.And.Prefix != nil {
						resf0elemf2f0.Prefix = resf0iter.Filter.And.Prefix
					}
					if resf0iter.Filter.And.Tags != nil {
						resf0elemf2f0f3 := []*svcapitypes.Tag{}
						for _, resf0elemf2f0f3iter := range resf0iter.Filter.And.Tags {
							resf0elemf2f0f3elem := &svcapitypes.Tag{}
							if resf0elemf2f0f3iter.Key != nil {
								resf0elemf2f0f3elem.Key = resf0elemf2f0f3iter.Key
							}
							if resf0elemf2f0f3iter.Value != nil {
								resf0elemf2f0f3elem.Value = resf0elemf2f0f3iter.Value
							}
							resf0elemf2f0f3 = append(resf0elemf2f0f3, resf0elemf2f0f3elem)
						}
						resf0elemf2f0.Tags = resf0elemf2f0f3
					}
					resf0elemf2.And = resf0elemf2f0
				}
				if resf0iter.Filter.ObjectSizeGreaterThan != nil {
					resf0elemf2.ObjectSizeGreaterThan = resf0iter.Filter.ObjectSizeGreaterThan
				}
				if resf0iter.Filter.ObjectSizeLessThan != nil {
					resf0elemf2.ObjectSizeLessThan = resf0iter.Filter.ObjectSizeLessThan
				}
				if resf0iter.Filter.Prefix != nil {
					resf0elemf2.Prefix = resf0iter.Filter.Prefix
				}
				if resf0iter.Filter.Tag != nil {
					resf0elemf2f4 := &svcapitypes.Tag{}
					if resf0iter.Filter.Tag.Key != nil {
						resf0elemf2f4.Key = resf0iter.Filter.Tag.Key
					}
					if resf0iter.Filter.Tag.Value != nil {
						resf0elemf2f4.Value = resf0iter.Filter.Tag.Value
					}
					resf0elemf2.Tag = resf0elemf2f4
				}
				resf0elem.Filter = resf0elemf2
			}
			if resf0iter.ID != nil {
				resf0elem.ID = resf0iter.ID
			}
			if resf0iter.NoncurrentVersionExpiration != nil {
				resf0elemf4 := &svcapitypes.NoncurrentVersionExpiration{}
				if resf0iter.NoncurrentVersionExpiration.NewerNoncurrentVersions != nil {
					newerNoncurrentVersionsCopy := int64(*resf0iter.NoncurrentVersionExpiration.NewerNoncurrentVersions)
					resf0elemf4.NewerNoncurrentVersions = &newerNoncurrentVersionsCopy
				}
				if resf0iter.NoncurrentVersionExpiration.NoncurrentDays != nil {
					noncurrentDaysCopy := int64(*resf0iter.NoncurrentVersionExpiration.NoncurrentDays)
					resf0elemf4.NoncurrentDays = &noncurrentDaysCopy
				}
				resf0elem.NoncurrentVersionExpiration = resf0elemf4
			}
			if resf0iter.NoncurrentVersionTransitions != nil {
				resf0elemf5 := []*svcapitypes.NoncurrentVersionTransition{}
				for _, resf0elemf5iter := range resf0iter.NoncurrentVersionTransitions {
					resf0elemf5elem := &svcapitypes.NoncurrentVersionTransition{}
					if resf0elemf5iter.NewerNoncurrentVersions != nil {
						newerNoncurrentVersionsCopy := int64(*resf0elemf5iter.NewerNoncurrentVersions)
						resf0elemf5elem.NewerNoncurrentVersions = &newerNoncurrentVersionsCopy
					}
					if resf0elemf5iter.NoncurrentDays != nil {
						noncurrentDaysCopy := int64(*resf0elemf5iter.NoncurrentDays)
						resf0elemf5elem.NoncurrentDays = &noncurrentDaysCopy
					}
					if resf0elemf5iter.StorageClass != "" {
						resf0elemf5elem.StorageClass = aws.String(string(resf0elemf5iter.StorageClass))
					}
					resf0elemf5 = append(resf0elemf5, resf0elemf5elem)
				}
				resf0elem.NoncurrentVersionTransitions = resf0elemf5
			}
			if resf0iter.Prefix != nil {
				resf0elem.Prefix = resf0iter.Prefix
			}
			if resf0iter.Status != "" {
				resf0elem.Status = aws.String(string(resf0iter.Status))
			}
			if resf0iter.Transitions != nil {
				resf0elemf8 := []*svcapitypes.Transition{}
				for _, resf0elemf8iter := range resf0iter.Transitions {
					resf0elemf8elem := &svcapitypes.Transition{}
					if resf0elemf8iter.Date != nil {
						resf0elemf8elem.Date = &metav1.Time{*resf0elemf8iter.Date}
					}
					if resf0elemf8iter.Days != nil {
						daysCopy := int64(*resf0elemf8iter.Days)
						resf0elemf8elem.Days = &daysCopy
					}
					if resf0elemf8iter.StorageClass != "" {
						resf0elemf8elem.StorageClass = aws.String(string(resf0elemf8iter.StorageClass))
					}
					resf0elemf8 = append(resf0elemf8, resf0elemf8elem)
				}
				resf0elem.Transitions = resf0elemf8
			}
			resf0 = append(resf0, resf0elem)
		}
		res.Rules = resf0
	}

	return res
}

// newBucketLoggingStatus returns a BucketLoggingStatus object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newBucketLoggingStatus(
	r *resource,
) *svcsdktypes.BucketLoggingStatus {
	res := &svcsdktypes.BucketLoggingStatus{}

	if r.ko.Spec.Logging.LoggingEnabled != nil {
		resf0 := &svcsdktypes.LoggingEnabled{}
		if r.ko.Spec.Logging.LoggingEnabled.TargetBucket != nil {
			resf0.TargetBucket = r.ko.Spec.Logging.LoggingEnabled.TargetBucket
		}
		if r.ko.Spec.Logging.LoggingEnabled.TargetGrants != nil {
			resf0f1 := []svcsdktypes.TargetGrant{}
			for _, resf0f1iter := range r.ko.Spec.Logging.LoggingEnabled.TargetGrants {
				resf0f1elem := &svcsdktypes.TargetGrant{}
				if resf0f1iter.Grantee != nil {
					resf0f1elemf0 := &svcsdktypes.Grantee{}
					if resf0f1iter.Grantee.DisplayName != nil {
						resf0f1elemf0.DisplayName = resf0f1iter.Grantee.DisplayName
					}
					if resf0f1iter.Grantee.EmailAddress != nil {
						resf0f1elemf0.EmailAddress = resf0f1iter.Grantee.EmailAddress
					}
					if resf0f1iter.Grantee.ID != nil {
						resf0f1elemf0.ID = resf0f1iter.Grantee.ID
					}
					if resf0f1iter.Grantee.Type != nil {
						resf0f1elemf0.Type = svcsdktypes.Type(*resf0f1iter.Grantee.Type)
					}
					if resf0f1iter.Grantee.URI != nil {
						resf0f1elemf0.URI = resf0f1iter.Grantee.URI
					}
					resf0f1elem.Grantee = resf0f1elemf0
				}
				if resf0f1iter.Permission != nil {
					resf0f1elem.Permission = svcsdktypes.BucketLogsPermission(*resf0f1iter.Permission)
				}
				resf0f1 = append(resf0f1, *resf0f1elem)
			}
			resf0.TargetGrants = resf0f1
		}
		if r.ko.Spec.Logging.LoggingEnabled.TargetPrefix != nil {
			resf0.TargetPrefix = r.ko.Spec.Logging.LoggingEnabled.TargetPrefix
		}
		res.LoggingEnabled = resf0
	}

	return res
}

// setResourceLogging sets the `Logging` spec field
// given the output of a `GetBucketLogging` operation.
func (rm *resourceManager) setResourceLogging(
	r *resource,
	resp *svcsdk.GetBucketLoggingOutput,
) *svcapitypes.BucketLoggingStatus {
	res := &svcapitypes.BucketLoggingStatus{}

	if resp.LoggingEnabled != nil {
		resf0 := &svcapitypes.LoggingEnabled{}
		if resp.LoggingEnabled.TargetBucket != nil {
			resf0.TargetBucket = resp.LoggingEnabled.TargetBucket
		}
		if resp.LoggingEnabled.TargetGrants != nil {
			resf0f1 := []*svcapitypes.TargetGrant{}
			for _, resf0f1iter := range resp.LoggingEnabled.TargetGrants {
				resf0f1elem := &svcapitypes.TargetGrant{}
				if resf0f1iter.Grantee != nil {
					resf0f1elemf0 := &svcapitypes.Grantee{}
					if resf0f1iter.Grantee.DisplayName != nil {
						resf0f1elemf0.DisplayName = resf0f1iter.Grantee.DisplayName
					}
					if resf0f1iter.Grantee.EmailAddress != nil {
						resf0f1elemf0.EmailAddress = resf0f1iter.Grantee.EmailAddress
					}
					if resf0f1iter.Grantee.ID != nil {
						resf0f1elemf0.ID = resf0f1iter.Grantee.ID
					}
					if resf0f1iter.Grantee.Type != "" {
						resf0f1elemf0.Type = aws.String(string(resf0f1iter.Grantee.Type))
					}
					if resf0f1iter.Grantee.URI != nil {
						resf0f1elemf0.URI = resf0f1iter.Grantee.URI
					}
					resf0f1elem.Grantee = resf0f1elemf0
				}
				if resf0f1iter.Permission != "" {
					resf0f1elem.Permission = aws.String(string(resf0f1iter.Permission))
				}
				resf0f1 = append(resf0f1, resf0f1elem)
			}
			resf0.TargetGrants = resf0f1
		}
		if resp.LoggingEnabled.TargetPrefix != nil {
			resf0.TargetPrefix = resp.LoggingEnabled.TargetPrefix
		}
		res.LoggingEnabled = resf0
	}

	return res
}

// newMetricsConfiguration returns a MetricsConfiguration object
// with each the field set by the corresponding configuration's fields.
func (rm *resourceManager) newMetricsConfiguration(
	c *svcapitypes.MetricsConfiguration,
) *svcsdktypes.MetricsConfiguration {
	res := &svcsdktypes.MetricsConfiguration{}

	if c.Filter != nil {
		if c.Filter.AccessPointARN != nil {
			resf0 := &svcsdktypes.MetricsFilterMemberAccessPointArn{}
			resf0.Value = *c.Filter.AccessPointARN
			res.Filter = resf0
		}
		if c.Filter.And != nil {
			resf0 := &svcsdktypes.MetricsFilterMemberAnd{}
			resf0f1 := svcsdktypes.MetricsAndOperator{}
			if c.Filter.And.AccessPointARN != nil {
				resf0f1.AccessPointArn = c.Filter.And.AccessPointARN
			}
			if c.Filter.And.Prefix != nil {
				resf0f1.Prefix = c.Filter.And.Prefix
			}
			if c.Filter.And.Tags != nil {
				resf0f1f2 := []svcsdktypes.Tag{}
				for _, resf0f1f2iter := range c.Filter.And.Tags {
					resf0f1f2elem := &svcsdktypes.Tag{}
					if resf0f1f2iter.Key != nil {
						resf0f1f2elem.Key = resf0f1f2iter.Key
					}
					if resf0f1f2iter.Value != nil {
						resf0f1f2elem.Value = resf0f1f2iter.Value
					}
					resf0f1f2 = append(resf0f1f2, *resf0f1f2elem)
				}
				resf0f1.Tags = resf0f1f2
			}
			resf0.Value = resf0f1
			res.Filter = resf0
		}
		if c.Filter.Prefix != nil {
			resf0 := &svcsdktypes.MetricsFilterMemberPrefix{}
			resf0.Value = *c.Filter.Prefix
			res.Filter = resf0
		}
		if c.Filter.Tag != nil {
			resf0 := &svcsdktypes.MetricsFilterMemberTag{}
			resf0f3 := svcsdktypes.Tag{}
			if c.Filter.Tag.Key != nil {
				resf0f3.Key = c.Filter.Tag.Key
			}
			if c.Filter.Tag.Value != nil {
				resf0f3.Value = c.Filter.Tag.Value
			}
			resf0.Value = resf0f3
			res.Filter = resf0
		}
	}
	if c.ID != nil {
		res.Id = c.ID
	}

	return res
}

// setMetricsConfiguration sets a resource MetricsConfiguration type
// given the SDK type.
func (rm *resourceManager) setResourceMetricsConfiguration(
	r *resource,
	resp *svcsdktypes.MetricsConfiguration,
) *svcapitypes.MetricsConfiguration {
	res := &svcapitypes.MetricsConfiguration{}

	if resp.Filter != nil {
		resf0 := &svcapitypes.MetricsFilter{}
		switch resp.Filter.(type) {
		case *svcsdktypes.MetricsFilterMemberAccessPointArn:
			f0 := resp.Filter.(*svcsdktypes.MetricsFilterMemberAccessPointArn)
			resf0.AccessPointARN = &f0.Value
		case *svcsdktypes.MetricsFilterMemberPrefix:
			f0 := resp.Filter.(*svcsdktypes.MetricsFilterMemberPrefix)
			resf0.Prefix = &f0.Value
		case *svcsdktypes.MetricsFilterMemberAnd:
			f0 := resp.Filter.(*svcsdktypes.MetricsFilterMemberAnd)
			if f0 != nil {
				resf0f1 := &svcapitypes.MetricsAndOperator{}
				resf0f1.Prefix = f0.Value.Prefix
				resf0f1.AccessPointARN = f0.Value.AccessPointArn
				resf0f1f2 := []*svcapitypes.Tag{}
				for _, resf0f1f2iter := range f0.Value.Tags {
					resf0f1f2elem := &svcapitypes.Tag{}
					if resf0f1f2iter.Key != nil {
						resf0f1f2elem.Key = resf0f1f2iter.Key
					}
					if resf0f1f2iter.Value != nil {
						resf0f1f2elem.Value = resf0f1f2iter.Value
					}
					resf0f1f2 = append(resf0f1f2, resf0f1f2elem)
				}
				resf0f1.Tags = resf0f1f2
			}
		}
		res.Filter = resf0
	}
	if resp.Id != nil {
		res.ID = resp.Id
	}

	return res
}

func compareMetricsConfiguration(
	a *svcapitypes.MetricsConfiguration,
	b *svcapitypes.MetricsConfiguration,
) *ackcompare.Delta {
	delta := ackcompare.NewDelta()
	if ackcompare.HasNilDifference(a.Filter, b.Filter) {
		delta.Add("MetricsConfiguration.Filter", a.Filter, b.Filter)
	} else if a.Filter != nil && b.Filter != nil {
		if ackcompare.HasNilDifference(a.Filter.AccessPointARN, b.Filter.AccessPointARN) {
			delta.Add("MetricsConfiguration.Filter.AccessPointARN", a.Filter.AccessPointARN, b.Filter.AccessPointARN)
		} else if a.Filter.AccessPointARN != nil && b.Filter.AccessPointARN != nil {
			if *a.Filter.AccessPointARN != *b.Filter.AccessPointARN {
				delta.Add("MetricsConfiguration.Filter.AccessPointARN", a.Filter.AccessPointARN, b.Filter.AccessPointARN)
			}
		}
		if ackcompare.HasNilDifference(a.Filter.And, b.Filter.And) {
			delta.Add("MetricsConfiguration.Filter.And", a.Filter.And, b.Filter.And)
		} else if a.Filter.And != nil && b.Filter.And != nil {
			if ackcompare.HasNilDifference(a.Filter.And.AccessPointARN, b.Filter.And.AccessPointARN) {
				delta.Add("MetricsConfiguration.Filter.And.AccessPointARN", a.Filter.And.AccessPointARN, b.Filter.And.AccessPointARN)
			} else if a.Filter.And.AccessPointARN != nil && b.Filter.And.AccessPointARN != nil {
				if *a.Filter.And.AccessPointARN != *b.Filter.And.AccessPointARN {
					delta.Add("MetricsConfiguration.Filter.And.AccessPointARN", a.Filter.And.AccessPointARN, b.Filter.And.AccessPointARN)
				}
			}
			if ackcompare.HasNilDifference(a.Filter.And.Prefix, b.Filter.And.Prefix) {
				delta.Add("MetricsConfiguration.Filter.And.Prefix", a.Filter.And.Prefix, b.Filter.And.Prefix)
			} else if a.Filter.And.Prefix != nil && b.Filter.And.Prefix != nil {
				if *a.Filter.And.Prefix != *b.Filter.And.Prefix {
					delta.Add("MetricsConfiguration.Filter.And.Prefix", a.Filter.And.Prefix, b.Filter.And.Prefix)
				}
			}
			if len(a.Filter.And.Tags) != len(b.Filter.And.Tags) {
				delta.Add("MetricsConfiguration.Filter.And.Tags", a.Filter.And.Tags, b.Filter.And.Tags)
			} else if len(a.Filter.And.Tags) > 0 {
				if !reflect.DeepEqual(a.Filter.And.Tags, b.Filter.And.Tags) {
					delta.Add("MetricsConfiguration.Filter.And.Tags", a.Filter.And.Tags, b.Filter.And.Tags)
				}
			}
		}
		if ackcompare.HasNilDifference(a.Filter.Prefix, b.Filter.Prefix) {
			delta.Add("MetricsConfiguration.Filter.Prefix", a.Filter.Prefix, b.Filter.Prefix)
		} else if a.Filter.Prefix != nil && b.Filter.Prefix != nil {
			if *a.Filter.Prefix != *b.Filter.Prefix {
				delta.Add("MetricsConfiguration.Filter.Prefix", a.Filter.Prefix, b.Filter.Prefix)
			}
		}
		if ackcompare.HasNilDifference(a.Filter.Tag, b.Filter.Tag) {
			delta.Add("MetricsConfiguration.Filter.Tag", a.Filter.Tag, b.Filter.Tag)
		} else if a.Filter.Tag != nil && b.Filter.Tag != nil {
			if ackcompare.HasNilDifference(a.Filter.Tag.Key, b.Filter.Tag.Key) {
				delta.Add("MetricsConfiguration.Filter.Tag.Key", a.Filter.Tag.Key, b.Filter.Tag.Key)
			} else if a.Filter.Tag.Key != nil && b.Filter.Tag.Key != nil {
				if *a.Filter.Tag.Key != *b.Filter.Tag.Key {
					delta.Add("MetricsConfiguration.Filter.Tag.Key", a.Filter.Tag.Key, b.Filter.Tag.Key)
				}
			}
			if ackcompare.HasNilDifference(a.Filter.Tag.Value, b.Filter.Tag.Value) {
				delta.Add("MetricsConfiguration.Filter.Tag.Value", a.Filter.Tag.Value, b.Filter.Tag.Value)
			} else if a.Filter.Tag.Value != nil && b.Filter.Tag.Value != nil {
				if *a.Filter.Tag.Value != *b.Filter.Tag.Value {
					delta.Add("MetricsConfiguration.Filter.Tag.Value", a.Filter.Tag.Value, b.Filter.Tag.Value)
				}
			}
		}
	}
	if ackcompare.HasNilDifference(a.ID, b.ID) {
		delta.Add("MetricsConfiguration.ID", a.ID, b.ID)
	} else if a.ID != nil && b.ID != nil {
		if *a.ID != *b.ID {
			delta.Add("MetricsConfiguration.ID", a.ID, b.ID)
		}
	}

	return delta
}

// getMetricsConfigurationAction returns the determined action for a given
// configuration object, depending on the desired and latest values
func getMetricsConfigurationAction(
	c *svcapitypes.MetricsConfiguration,
	latest *resource,
) ConfigurationAction {
	action := ConfigurationActionPut
	if latest != nil {
		for _, l := range latest.ko.Spec.Metrics {
			if *l.ID != *c.ID {
				continue
			}

			// Don't perform any action if they are identical
			delta := compareMetricsConfiguration(l, c)
			if len(delta.Differences) > 0 {
				action = ConfigurationActionUpdate
			} else {
				action = ConfigurationActionNone
			}
			break
		}
	}
	return action
}

func (rm *resourceManager) newListBucketMetricsPayload(
	r *resource,
) *svcsdk.ListBucketMetricsConfigurationsInput {
	res := &svcsdk.ListBucketMetricsConfigurationsInput{}
	res.Bucket = r.ko.Spec.Name
	return res
}

func (rm *resourceManager) newPutBucketMetricsPayload(
	r *resource,
	c svcapitypes.MetricsConfiguration,
) *svcsdk.PutBucketMetricsConfigurationInput {
	res := &svcsdk.PutBucketMetricsConfigurationInput{}
	res.Bucket = r.ko.Spec.Name
	res.Id = c.ID
	res.MetricsConfiguration = rm.newMetricsConfiguration(&c)

	return res
}

func (rm *resourceManager) newDeleteBucketMetricsPayload(
	r *resource,
	c svcapitypes.MetricsConfiguration,
) *svcsdk.DeleteBucketMetricsConfigurationInput {
	res := &svcsdk.DeleteBucketMetricsConfigurationInput{}
	res.Bucket = r.ko.Spec.Name
	res.Id = c.ID

	return res
}

func (rm *resourceManager) deleteMetricsConfiguration(
	ctx context.Context,
	r *resource,
	c svcapitypes.MetricsConfiguration,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.deleteMetricsConfiguration")
	defer exit(err)

	input := rm.newDeleteBucketMetricsPayload(r, c)
	_, err = rm.sdkapi.DeleteBucketMetricsConfiguration(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeleteBucketMetricsConfiguration", err)
	return err
}

func (rm *resourceManager) putMetricsConfiguration(
	ctx context.Context,
	r *resource,
	c svcapitypes.MetricsConfiguration,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.putMetricsConfiguration")
	defer exit(err)

	input := rm.newPutBucketMetricsPayload(r, c)
	_, err = rm.sdkapi.PutBucketMetricsConfiguration(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucketMetricsConfiguration", err)
	return err
}

func (rm *resourceManager) syncMetrics(
	ctx context.Context,
	desired *resource,
	latest *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncMetrics")
	defer exit(err)

	for _, c := range desired.ko.Spec.Metrics {
		action := getMetricsConfigurationAction(c, latest)

		switch action {
		case ConfigurationActionUpdate:
			fallthrough
		case ConfigurationActionPut:
			if err = rm.putMetricsConfiguration(ctx, desired, *c); err != nil {
				return err
			}
		default:
		}
	}

	if latest != nil {
		// Find any configurations that are in the latest but not in desired
		for _, l := range latest.ko.Spec.Metrics {
			exists := false
			for _, c := range desired.ko.Spec.Metrics {
				if *c.ID != *l.ID {
					continue
				}
				exists = true
				break
			}

			if !exists {
				if err = rm.deleteMetricsConfiguration(ctx, desired, *l); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// newNotificationConfiguration returns a NotificationConfiguration object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newNotificationConfiguration(
	r *resource,
) *svcsdktypes.NotificationConfiguration {
	res := &svcsdktypes.NotificationConfiguration{}

	if r.ko.Spec.Notification.LambdaFunctionConfigurations != nil {
		resf0 := []svcsdktypes.LambdaFunctionConfiguration{}
		for _, resf0iter := range r.ko.Spec.Notification.LambdaFunctionConfigurations {
			resf0elem := &svcsdktypes.LambdaFunctionConfiguration{}
			if resf0iter.Events != nil {
				resf0elemf0 := []svcsdktypes.Event{}
				for _, resf0elemf0iter := range resf0iter.Events {
					var resf0elemf0elem string
					resf0elemf0elem = string(*resf0elemf0iter)
					resf0elemf0 = append(resf0elemf0, svcsdktypes.Event(resf0elemf0elem))
				}
				resf0elem.Events = resf0elemf0
			}
			if resf0iter.Filter != nil {
				resf0elemf1 := &svcsdktypes.NotificationConfigurationFilter{}
				if resf0iter.Filter.Key != nil {
					resf0elemf1f0 := &svcsdktypes.S3KeyFilter{}
					if resf0iter.Filter.Key.FilterRules != nil {
						resf0elemf1f0f0 := []svcsdktypes.FilterRule{}
						for _, resf0elemf1f0f0iter := range resf0iter.Filter.Key.FilterRules {
							resf0elemf1f0f0elem := &svcsdktypes.FilterRule{}
							if resf0elemf1f0f0iter.Name != nil {
								resf0elemf1f0f0elem.Name = svcsdktypes.FilterRuleName(*resf0elemf1f0f0iter.Name)
							}
							if resf0elemf1f0f0iter.Value != nil {
								resf0elemf1f0f0elem.Value = resf0elemf1f0f0iter.Value
							}
							resf0elemf1f0f0 = append(resf0elemf1f0f0, *resf0elemf1f0f0elem)
						}
						resf0elemf1f0.FilterRules = resf0elemf1f0f0
					}
					resf0elemf1.Key = resf0elemf1f0
				}
				resf0elem.Filter = resf0elemf1
			}
			if resf0iter.ID != nil {
				resf0elem.Id = resf0iter.ID
			}
			if resf0iter.LambdaFunctionARN != nil {
				resf0elem.LambdaFunctionArn = resf0iter.LambdaFunctionARN
			}
			resf0 = append(resf0, *resf0elem)
		}
		res.LambdaFunctionConfigurations = resf0
	}
	if r.ko.Spec.Notification.QueueConfigurations != nil {
		resf1 := []svcsdktypes.QueueConfiguration{}
		for _, resf1iter := range r.ko.Spec.Notification.QueueConfigurations {
			resf1elem := &svcsdktypes.QueueConfiguration{}
			if resf1iter.Events != nil {
				resf1elemf0 := []svcsdktypes.Event{}
				for _, resf1elemf0iter := range resf1iter.Events {
					var resf1elemf0elem string
					resf1elemf0elem = string(*resf1elemf0iter)
					resf1elemf0 = append(resf1elemf0, svcsdktypes.Event(resf1elemf0elem))
				}
				resf1elem.Events = resf1elemf0
			}
			if resf1iter.Filter != nil {
				resf1elemf1 := &svcsdktypes.NotificationConfigurationFilter{}
				if resf1iter.Filter.Key != nil {
					resf1elemf1f0 := &svcsdktypes.S3KeyFilter{}
					if resf1iter.Filter.Key.FilterRules != nil {
						resf1elemf1f0f0 := []svcsdktypes.FilterRule{}
						for _, resf1elemf1f0f0iter := range resf1iter.Filter.Key.FilterRules {
							resf1elemf1f0f0elem := &svcsdktypes.FilterRule{}
							if resf1elemf1f0f0iter.Name != nil {
								resf1elemf1f0f0elem.Name = svcsdktypes.FilterRuleName(*resf1elemf1f0f0iter.Name)
							}
							if resf1elemf1f0f0iter.Value != nil {
								resf1elemf1f0f0elem.Value = resf1elemf1f0f0iter.Value
							}
							resf1elemf1f0f0 = append(resf1elemf1f0f0, *resf1elemf1f0f0elem)
						}
						resf1elemf1f0.FilterRules = resf1elemf1f0f0
					}
					resf1elemf1.Key = resf1elemf1f0
				}
				resf1elem.Filter = resf1elemf1
			}
			if resf1iter.ID != nil {
				resf1elem.Id = resf1iter.ID
			}
			if resf1iter.QueueARN != nil {
				resf1elem.QueueArn = resf1iter.QueueARN
			}
			resf1 = append(resf1, *resf1elem)
		}
		res.QueueConfigurations = resf1
	}
	if r.ko.Spec.Notification.TopicConfigurations != nil {
		resf2 := []svcsdktypes.TopicConfiguration{}
		for _, resf2iter := range r.ko.Spec.Notification.TopicConfigurations {
			resf2elem := &svcsdktypes.TopicConfiguration{}
			if resf2iter.Events != nil {
				resf2elemf0 := []svcsdktypes.Event{}
				for _, resf2elemf0iter := range resf2iter.Events {
					var resf2elemf0elem string
					resf2elemf0elem = string(*resf2elemf0iter)
					resf2elemf0 = append(resf2elemf0, svcsdktypes.Event(resf2elemf0elem))
				}
				resf2elem.Events = resf2elemf0
			}
			if resf2iter.Filter != nil {
				resf2elemf1 := &svcsdktypes.NotificationConfigurationFilter{}
				if resf2iter.Filter.Key != nil {
					resf2elemf1f0 := &svcsdktypes.S3KeyFilter{}
					if resf2iter.Filter.Key.FilterRules != nil {
						resf2elemf1f0f0 := []svcsdktypes.FilterRule{}
						for _, resf2elemf1f0f0iter := range resf2iter.Filter.Key.FilterRules {
							resf2elemf1f0f0elem := &svcsdktypes.FilterRule{}
							if resf2elemf1f0f0iter.Name != nil {
								resf2elemf1f0f0elem.Name = svcsdktypes.FilterRuleName(*resf2elemf1f0f0iter.Name)
							}
							if resf2elemf1f0f0iter.Value != nil {
								resf2elemf1f0f0elem.Value = resf2elemf1f0f0iter.Value
							}
							resf2elemf1f0f0 = append(resf2elemf1f0f0, *resf2elemf1f0f0elem)
						}
						resf2elemf1f0.FilterRules = resf2elemf1f0f0
					}
					resf2elemf1.Key = resf2elemf1f0
				}
				resf2elem.Filter = resf2elemf1
			}
			if resf2iter.ID != nil {
				resf2elem.Id = resf2iter.ID
			}
			if resf2iter.TopicARN != nil {
				resf2elem.TopicArn = resf2iter.TopicARN
			}
			resf2 = append(resf2, *resf2elem)
		}
		res.TopicConfigurations = resf2
	}

	return res
}

// setResourceNotification sets the `Notification` spec field
// given the output of a `GetBucketNotificationConfiguration` operation.
func (rm *resourceManager) setResourceNotification(
	r *resource,
	resp *svcsdk.GetBucketNotificationConfigurationOutput,
) *svcapitypes.NotificationConfiguration {
	res := &svcapitypes.NotificationConfiguration{}

	if resp.LambdaFunctionConfigurations != nil {
		resf0 := []*svcapitypes.LambdaFunctionConfiguration{}
		for _, resf0iter := range resp.LambdaFunctionConfigurations {
			resf0elem := &svcapitypes.LambdaFunctionConfiguration{}
			if resf0iter.Events != nil {
				resf0elemf0 := []*string{}
				for _, resf0elemf0iter := range resf0iter.Events {
					var resf0elemf0elem *string
					resf0elemf0elem = aws.String(string(resf0elemf0iter))
					resf0elemf0 = append(resf0elemf0, resf0elemf0elem)
				}
				resf0elem.Events = resf0elemf0
			}
			if resf0iter.Filter != nil {
				resf0elemf1 := &svcapitypes.NotificationConfigurationFilter{}
				if resf0iter.Filter.Key != nil {
					resf0elemf1f0 := &svcapitypes.KeyFilter{}
					if resf0iter.Filter.Key.FilterRules != nil {
						resf0elemf1f0f0 := []*svcapitypes.FilterRule{}
						for _, resf0elemf1f0f0iter := range resf0iter.Filter.Key.FilterRules {
							resf0elemf1f0f0elem := &svcapitypes.FilterRule{}
							if resf0elemf1f0f0iter.Name != "" {
								resf0elemf1f0f0elem.Name = aws.String(string(resf0elemf1f0f0iter.Name))
							}
							if resf0elemf1f0f0iter.Value != nil {
								resf0elemf1f0f0elem.Value = resf0elemf1f0f0iter.Value
							}
							resf0elemf1f0f0 = append(resf0elemf1f0f0, resf0elemf1f0f0elem)
						}
						resf0elemf1f0.FilterRules = resf0elemf1f0f0
					}
					resf0elemf1.Key = resf0elemf1f0
				}
				resf0elem.Filter = resf0elemf1
			}
			if resf0iter.Id != nil {
				resf0elem.ID = resf0iter.Id
			}
			if resf0iter.LambdaFunctionArn != nil {
				resf0elem.LambdaFunctionARN = resf0iter.LambdaFunctionArn
			}
			resf0 = append(resf0, resf0elem)
		}
		res.LambdaFunctionConfigurations = resf0
	}
	if resp.QueueConfigurations != nil {
		resf1 := []*svcapitypes.QueueConfiguration{}
		for _, resf1iter := range resp.QueueConfigurations {
			resf1elem := &svcapitypes.QueueConfiguration{}
			if resf1iter.Events != nil {
				resf1elemf0 := []*string{}
				for _, resf1elemf0iter := range resf1iter.Events {
					var resf1elemf0elem *string
					resf1elemf0elem = aws.String(string(resf1elemf0iter))
					resf1elemf0 = append(resf1elemf0, resf1elemf0elem)
				}
				resf1elem.Events = resf1elemf0
			}
			if resf1iter.Filter != nil {
				resf1elemf1 := &svcapitypes.NotificationConfigurationFilter{}
				if resf1iter.Filter.Key != nil {
					resf1elemf1f0 := &svcapitypes.KeyFilter{}
					if resf1iter.Filter.Key.FilterRules != nil {
						resf1elemf1f0f0 := []*svcapitypes.FilterRule{}
						for _, resf1elemf1f0f0iter := range resf1iter.Filter.Key.FilterRules {
							resf1elemf1f0f0elem := &svcapitypes.FilterRule{}
							if resf1elemf1f0f0iter.Name != "" {
								resf1elemf1f0f0elem.Name = aws.String(string(resf1elemf1f0f0iter.Name))
							}
							if resf1elemf1f0f0iter.Value != nil {
								resf1elemf1f0f0elem.Value = resf1elemf1f0f0iter.Value
							}
							resf1elemf1f0f0 = append(resf1elemf1f0f0, resf1elemf1f0f0elem)
						}
						resf1elemf1f0.FilterRules = resf1elemf1f0f0
					}
					resf1elemf1.Key = resf1elemf1f0
				}
				resf1elem.Filter = resf1elemf1
			}
			if resf1iter.Id != nil {
				resf1elem.ID = resf1iter.Id
			}
			if resf1iter.QueueArn != nil {
				resf1elem.QueueARN = resf1iter.QueueArn
			}
			resf1 = append(resf1, resf1elem)
		}
		res.QueueConfigurations = resf1
	}
	if resp.TopicConfigurations != nil {
		resf2 := []*svcapitypes.TopicConfiguration{}
		for _, resf2iter := range resp.TopicConfigurations {
			resf2elem := &svcapitypes.TopicConfiguration{}
			if resf2iter.Events != nil {
				resf2elemf0 := []*string{}
				for _, resf2elemf0iter := range resf2iter.Events {
					var resf2elemf0elem *string
					resf2elemf0elem = aws.String(string(resf2elemf0iter))
					resf2elemf0 = append(resf2elemf0, resf2elemf0elem)
				}
				resf2elem.Events = resf2elemf0
			}
			if resf2iter.Filter != nil {
				resf2elemf1 := &svcapitypes.NotificationConfigurationFilter{}
				if resf2iter.Filter.Key != nil {
					resf2elemf1f0 := &svcapitypes.KeyFilter{}
					if resf2iter.Filter.Key.FilterRules != nil {
						resf2elemf1f0f0 := []*svcapitypes.FilterRule{}
						for _, resf2elemf1f0f0iter := range resf2iter.Filter.Key.FilterRules {
							resf2elemf1f0f0elem := &svcapitypes.FilterRule{}
							if resf2elemf1f0f0iter.Name != "" {
								resf2elemf1f0f0elem.Name = aws.String(string(resf2elemf1f0f0iter.Name))
							}
							if resf2elemf1f0f0iter.Value != nil {
								resf2elemf1f0f0elem.Value = resf2elemf1f0f0iter.Value
							}
							resf2elemf1f0f0 = append(resf2elemf1f0f0, resf2elemf1f0f0elem)
						}
						resf2elemf1f0.FilterRules = resf2elemf1f0f0
					}
					resf2elemf1.Key = resf2elemf1f0
				}
				resf2elem.Filter = resf2elemf1
			}
			if resf2iter.Id != nil {
				resf2elem.ID = resf2iter.Id
			}
			if resf2iter.TopicArn != nil {
				resf2elem.TopicARN = resf2iter.TopicArn
			}
			resf2 = append(resf2, resf2elem)
		}
		res.TopicConfigurations = resf2
	}

	return res
}

// newOwnershipControls returns a OwnershipControls object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newOwnershipControls(
	r *resource,
) *svcsdktypes.OwnershipControls {
	res := &svcsdktypes.OwnershipControls{}

	if r.ko.Spec.OwnershipControls.Rules != nil {
		resf0 := []svcsdktypes.OwnershipControlsRule{}
		for _, resf0iter := range r.ko.Spec.OwnershipControls.Rules {
			resf0elem := &svcsdktypes.OwnershipControlsRule{}
			if resf0iter.ObjectOwnership != nil {
				resf0elem.ObjectOwnership = svcsdktypes.ObjectOwnership(*resf0iter.ObjectOwnership)
			}
			resf0 = append(resf0, *resf0elem)
		}
		res.Rules = resf0
	}

	return res
}

// setResourceOwnershipControls sets the `OwnershipControls` spec field
// given the output of a `GetBucketOwnershipControls` operation.
func (rm *resourceManager) setResourceOwnershipControls(
	r *resource,
	resp *svcsdk.GetBucketOwnershipControlsOutput,
) *svcapitypes.OwnershipControls {
	res := &svcapitypes.OwnershipControls{}

	if resp.OwnershipControls.Rules != nil {
		resf0 := []*svcapitypes.OwnershipControlsRule{}
		for _, resf0iter := range resp.OwnershipControls.Rules {
			resf0elem := &svcapitypes.OwnershipControlsRule{}
			if resf0iter.ObjectOwnership != "" {
				resf0elem.ObjectOwnership = aws.String(string(resf0iter.ObjectOwnership))
			}
			resf0 = append(resf0, resf0elem)
		}
		res.Rules = resf0
	}

	return res
}

// newPublicAccessBlockConfiguration returns a PublicAccessBlockConfiguration object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newPublicAccessBlockConfiguration(
	r *resource,
) *svcsdktypes.PublicAccessBlockConfiguration {
	res := &svcsdktypes.PublicAccessBlockConfiguration{}

	if r.ko.Spec.PublicAccessBlock.BlockPublicACLs != nil {
		res.BlockPublicAcls = r.ko.Spec.PublicAccessBlock.BlockPublicACLs
	}
	if r.ko.Spec.PublicAccessBlock.BlockPublicPolicy != nil {
		res.BlockPublicPolicy = r.ko.Spec.PublicAccessBlock.BlockPublicPolicy
	}
	if r.ko.Spec.PublicAccessBlock.IgnorePublicACLs != nil {
		res.IgnorePublicAcls = r.ko.Spec.PublicAccessBlock.IgnorePublicACLs
	}
	if r.ko.Spec.PublicAccessBlock.RestrictPublicBuckets != nil {
		res.RestrictPublicBuckets = r.ko.Spec.PublicAccessBlock.RestrictPublicBuckets
	}

	return res
}

// setResourcePublicAccessBlock sets the `PublicAccessBlock` spec field
// given the output of a `GetPublicAccessBlock` operation.
func (rm *resourceManager) setResourcePublicAccessBlock(
	r *resource,
	resp *svcsdk.GetPublicAccessBlockOutput,
) *svcapitypes.PublicAccessBlockConfiguration {
	res := &svcapitypes.PublicAccessBlockConfiguration{}

	if resp.PublicAccessBlockConfiguration.BlockPublicAcls != nil {
		res.BlockPublicACLs = resp.PublicAccessBlockConfiguration.BlockPublicAcls
	}
	if resp.PublicAccessBlockConfiguration.BlockPublicPolicy != nil {
		res.BlockPublicPolicy = resp.PublicAccessBlockConfiguration.BlockPublicPolicy
	}
	if resp.PublicAccessBlockConfiguration.IgnorePublicAcls != nil {
		res.IgnorePublicACLs = resp.PublicAccessBlockConfiguration.IgnorePublicAcls
	}
	if resp.PublicAccessBlockConfiguration.RestrictPublicBuckets != nil {
		res.RestrictPublicBuckets = resp.PublicAccessBlockConfiguration.RestrictPublicBuckets
	}

	return res
}

// newReplicationConfiguration returns a ReplicationConfiguration object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newReplicationConfiguration(
	r *resource,
) *svcsdktypes.ReplicationConfiguration {
	res := &svcsdktypes.ReplicationConfiguration{}

	if r.ko.Spec.Replication.Role != nil {
		res.Role = r.ko.Spec.Replication.Role
	}
	if r.ko.Spec.Replication.Rules != nil {
		resf1 := []svcsdktypes.ReplicationRule{}
		for _, resf1iter := range r.ko.Spec.Replication.Rules {
			resf1elem := &svcsdktypes.ReplicationRule{}
			if resf1iter.DeleteMarkerReplication != nil {
				resf1elemf0 := &svcsdktypes.DeleteMarkerReplication{}
				if resf1iter.DeleteMarkerReplication.Status != nil {
					resf1elemf0.Status = svcsdktypes.DeleteMarkerReplicationStatus(*resf1iter.DeleteMarkerReplication.Status)
				}
				resf1elem.DeleteMarkerReplication = resf1elemf0
			}
			if resf1iter.Destination != nil {
				resf1elemf1 := &svcsdktypes.Destination{}
				if resf1iter.Destination.AccessControlTranslation != nil {
					resf1elemf1f0 := &svcsdktypes.AccessControlTranslation{}
					if resf1iter.Destination.AccessControlTranslation.Owner != nil {
						resf1elemf1f0.Owner = svcsdktypes.OwnerOverride(*resf1iter.Destination.AccessControlTranslation.Owner)
					}
					resf1elemf1.AccessControlTranslation = resf1elemf1f0
				}
				if resf1iter.Destination.Account != nil {
					resf1elemf1.Account = resf1iter.Destination.Account
				}
				if resf1iter.Destination.Bucket != nil {
					resf1elemf1.Bucket = resf1iter.Destination.Bucket
				}
				if resf1iter.Destination.EncryptionConfiguration != nil {
					resf1elemf1f3 := &svcsdktypes.EncryptionConfiguration{}
					if resf1iter.Destination.EncryptionConfiguration.ReplicaKMSKeyID != nil {
						resf1elemf1f3.ReplicaKmsKeyID = resf1iter.Destination.EncryptionConfiguration.ReplicaKMSKeyID
					}
					resf1elemf1.EncryptionConfiguration = resf1elemf1f3
				}
				if resf1iter.Destination.Metrics != nil {
					resf1elemf1f4 := &svcsdktypes.Metrics{}
					if resf1iter.Destination.Metrics.EventThreshold != nil {
						resf1elemf1f4f0 := &svcsdktypes.ReplicationTimeValue{}
						if resf1iter.Destination.Metrics.EventThreshold.Minutes != nil {
							minutesCopy := int32(*resf1iter.Destination.Metrics.EventThreshold.Minutes)
							resf1elemf1f4f0.Minutes = &minutesCopy
						}
						resf1elemf1f4.EventThreshold = resf1elemf1f4f0
					}
					if resf1iter.Destination.Metrics.Status != nil {
						resf1elemf1f4.Status = svcsdktypes.MetricsStatus(*resf1iter.Destination.Metrics.Status)
					}
					resf1elemf1.Metrics = resf1elemf1f4
				}
				if resf1iter.Destination.ReplicationTime != nil {
					resf1elemf1f5 := &svcsdktypes.ReplicationTime{}
					if resf1iter.Destination.ReplicationTime.Status != nil {
						resf1elemf1f5.Status = svcsdktypes.ReplicationTimeStatus(*resf1iter.Destination.ReplicationTime.Status)
					}
					if resf1iter.Destination.ReplicationTime.Time != nil {
						resf1elemf1f5f1 := &svcsdktypes.ReplicationTimeValue{}
						if resf1iter.Destination.ReplicationTime.Time.Minutes != nil {
							minutesCopy := int32(*resf1iter.Destination.ReplicationTime.Time.Minutes)
							resf1elemf1f5f1.Minutes = &minutesCopy
						}
						resf1elemf1f5.Time = resf1elemf1f5f1
					}
					resf1elemf1.ReplicationTime = resf1elemf1f5
				}
				if resf1iter.Destination.StorageClass != nil {
					resf1elemf1.StorageClass = svcsdktypes.StorageClass(*resf1iter.Destination.StorageClass)
				}
				resf1elem.Destination = resf1elemf1
			}
			if resf1iter.ExistingObjectReplication != nil {
				resf1elemf2 := &svcsdktypes.ExistingObjectReplication{}
				if resf1iter.ExistingObjectReplication.Status != nil {
					resf1elemf2.Status = svcsdktypes.ExistingObjectReplicationStatus(*resf1iter.ExistingObjectReplication.Status)
				}
				resf1elem.ExistingObjectReplication = resf1elemf2
			}
			if resf1iter.Filter != nil {
				resf1elemf3 := &svcsdktypes.ReplicationRuleFilter{}
				if resf1iter.Filter.And != nil {
					resf1elemf3f0 := &svcsdktypes.ReplicationRuleAndOperator{}
					if resf1iter.Filter.And.Prefix != nil {
						resf1elemf3f0.Prefix = resf1iter.Filter.And.Prefix
					}
					if resf1iter.Filter.And.Tags != nil {
						resf1elemf3f0f1 := []svcsdktypes.Tag{}
						for _, resf1elemf3f0f1iter := range resf1iter.Filter.And.Tags {
							resf1elemf3f0f1elem := &svcsdktypes.Tag{}
							if resf1elemf3f0f1iter.Key != nil {
								resf1elemf3f0f1elem.Key = resf1elemf3f0f1iter.Key
							}
							if resf1elemf3f0f1iter.Value != nil {
								resf1elemf3f0f1elem.Value = resf1elemf3f0f1iter.Value
							}
							resf1elemf3f0f1 = append(resf1elemf3f0f1, *resf1elemf3f0f1elem)
						}
						resf1elemf3f0.Tags = resf1elemf3f0f1
					}
					resf1elemf3.And = resf1elemf3f0
				}
				if resf1iter.Filter.Prefix != nil {
					resf1elemf3.Prefix = resf1iter.Filter.Prefix
				}
				if resf1iter.Filter.Tag != nil {
					resf1elemf3f2 := &svcsdktypes.Tag{}
					if resf1iter.Filter.Tag.Key != nil {
						resf1elemf3f2.Key = resf1iter.Filter.Tag.Key
					}
					if resf1iter.Filter.Tag.Value != nil {
						resf1elemf3f2.Value = resf1iter.Filter.Tag.Value
					}
					resf1elemf3.Tag = resf1elemf3f2
				}
				resf1elem.Filter = resf1elemf3
			}
			if resf1iter.ID != nil {
				resf1elem.ID = resf1iter.ID
			}
			if resf1iter.Prefix != nil {
				resf1elem.Prefix = resf1iter.Prefix
			}
			if resf1iter.Priority != nil {
				priorityCopy := int32(*resf1iter.Priority)
				resf1elem.Priority = &priorityCopy
			}
			if resf1iter.SourceSelectionCriteria != nil {
				resf1elemf7 := &svcsdktypes.SourceSelectionCriteria{}
				if resf1iter.SourceSelectionCriteria.ReplicaModifications != nil {
					resf1elemf7f0 := &svcsdktypes.ReplicaModifications{}
					if resf1iter.SourceSelectionCriteria.ReplicaModifications.Status != nil {
						resf1elemf7f0.Status = svcsdktypes.ReplicaModificationsStatus(*resf1iter.SourceSelectionCriteria.ReplicaModifications.Status)
					}
					resf1elemf7.ReplicaModifications = resf1elemf7f0
				}
				if resf1iter.SourceSelectionCriteria.SSEKMSEncryptedObjects != nil {
					resf1elemf7f1 := &svcsdktypes.SseKmsEncryptedObjects{}
					if resf1iter.SourceSelectionCriteria.SSEKMSEncryptedObjects.Status != nil {
						resf1elemf7f1.Status = svcsdktypes.SseKmsEncryptedObjectsStatus(*resf1iter.SourceSelectionCriteria.SSEKMSEncryptedObjects.Status)
					}
					resf1elemf7.SseKmsEncryptedObjects = resf1elemf7f1
				}
				resf1elem.SourceSelectionCriteria = resf1elemf7
			}
			if resf1iter.Status != nil {
				resf1elem.Status = svcsdktypes.ReplicationRuleStatus(*resf1iter.Status)
			}
			resf1 = append(resf1, *resf1elem)
		}
		res.Rules = resf1
	}

	return res
}

// setResourceReplication sets the `Replication` spec field
// given the output of a `GetBucketReplication` operation.
func (rm *resourceManager) setResourceReplication(
	r *resource,
	resp *svcsdk.GetBucketReplicationOutput,
) *svcapitypes.ReplicationConfiguration {
	res := &svcapitypes.ReplicationConfiguration{}

	if resp.ReplicationConfiguration.Role != nil {
		res.Role = resp.ReplicationConfiguration.Role
	}
	if resp.ReplicationConfiguration.Rules != nil {
		resf1 := []*svcapitypes.ReplicationRule{}
		for _, resf1iter := range resp.ReplicationConfiguration.Rules {
			resf1elem := &svcapitypes.ReplicationRule{}
			if resf1iter.DeleteMarkerReplication != nil {
				resf1elemf0 := &svcapitypes.DeleteMarkerReplication{}
				if resf1iter.DeleteMarkerReplication.Status != "" {
					resf1elemf0.Status = aws.String(string(resf1iter.DeleteMarkerReplication.Status))
				}
				resf1elem.DeleteMarkerReplication = resf1elemf0
			}
			if resf1iter.Destination != nil {
				resf1elemf1 := &svcapitypes.Destination{}
				if resf1iter.Destination.AccessControlTranslation != nil {
					resf1elemf1f0 := &svcapitypes.AccessControlTranslation{}
					if resf1iter.Destination.AccessControlTranslation.Owner != "" {
						resf1elemf1f0.Owner = aws.String(string(resf1iter.Destination.AccessControlTranslation.Owner))
					}
					resf1elemf1.AccessControlTranslation = resf1elemf1f0
				}
				if resf1iter.Destination.Account != nil {
					resf1elemf1.Account = resf1iter.Destination.Account
				}
				if resf1iter.Destination.Bucket != nil {
					resf1elemf1.Bucket = resf1iter.Destination.Bucket
				}
				if resf1iter.Destination.EncryptionConfiguration != nil {
					resf1elemf1f3 := &svcapitypes.EncryptionConfiguration{}
					if resf1iter.Destination.EncryptionConfiguration.ReplicaKmsKeyID != nil {
						resf1elemf1f3.ReplicaKMSKeyID = resf1iter.Destination.EncryptionConfiguration.ReplicaKmsKeyID
					}
					resf1elemf1.EncryptionConfiguration = resf1elemf1f3
				}
				if resf1iter.Destination.Metrics != nil {
					resf1elemf1f4 := &svcapitypes.Metrics{}
					if resf1iter.Destination.Metrics.EventThreshold != nil {
						resf1elemf1f4f0 := &svcapitypes.ReplicationTimeValue{}
						if resf1iter.Destination.Metrics.EventThreshold.Minutes != nil {
							minutesCopy := int64(*resf1iter.Destination.Metrics.EventThreshold.Minutes)
							resf1elemf1f4f0.Minutes = &minutesCopy
						}
						resf1elemf1f4.EventThreshold = resf1elemf1f4f0
					}
					if resf1iter.Destination.Metrics.Status != "" {
						resf1elemf1f4.Status = aws.String(string(resf1iter.Destination.Metrics.Status))
					}
					resf1elemf1.Metrics = resf1elemf1f4
				}
				if resf1iter.Destination.ReplicationTime != nil {
					resf1elemf1f5 := &svcapitypes.ReplicationTime{}
					if resf1iter.Destination.ReplicationTime.Status != "" {
						resf1elemf1f5.Status = aws.String(string(resf1iter.Destination.ReplicationTime.Status))
					}
					if resf1iter.Destination.ReplicationTime.Time != nil {
						resf1elemf1f5f1 := &svcapitypes.ReplicationTimeValue{}
						if resf1iter.Destination.ReplicationTime.Time.Minutes != nil {
							minutesCopy := int64(*resf1iter.Destination.ReplicationTime.Time.Minutes)
							resf1elemf1f5f1.Minutes = &minutesCopy
						}
						resf1elemf1f5.Time = resf1elemf1f5f1
					}
					resf1elemf1.ReplicationTime = resf1elemf1f5
				}
				if resf1iter.Destination.StorageClass != "" {
					resf1elemf1.StorageClass = aws.String(string(resf1iter.Destination.StorageClass))
				}
				resf1elem.Destination = resf1elemf1
			}
			if resf1iter.ExistingObjectReplication != nil {
				resf1elemf2 := &svcapitypes.ExistingObjectReplication{}
				if resf1iter.ExistingObjectReplication.Status != "" {
					resf1elemf2.Status = aws.String(string(resf1iter.ExistingObjectReplication.Status))
				}
				resf1elem.ExistingObjectReplication = resf1elemf2
			}
			if resf1iter.Filter != nil {
				resf1elemf3 := &svcapitypes.ReplicationRuleFilter{}
				if resf1iter.Filter.And != nil {
					resf1elemf3f0 := &svcapitypes.ReplicationRuleAndOperator{}
					if resf1iter.Filter.And.Prefix != nil {
						resf1elemf3f0.Prefix = resf1iter.Filter.And.Prefix
					}
					if resf1iter.Filter.And.Tags != nil {
						resf1elemf3f0f1 := []*svcapitypes.Tag{}
						for _, resf1elemf3f0f1iter := range resf1iter.Filter.And.Tags {
							resf1elemf3f0f1elem := &svcapitypes.Tag{}
							if resf1elemf3f0f1iter.Key != nil {
								resf1elemf3f0f1elem.Key = resf1elemf3f0f1iter.Key
							}
							if resf1elemf3f0f1iter.Value != nil {
								resf1elemf3f0f1elem.Value = resf1elemf3f0f1iter.Value
							}
							resf1elemf3f0f1 = append(resf1elemf3f0f1, resf1elemf3f0f1elem)
						}
						resf1elemf3f0.Tags = resf1elemf3f0f1
					}
					resf1elemf3.And = resf1elemf3f0
				}
				if resf1iter.Filter.Prefix != nil {
					resf1elemf3.Prefix = resf1iter.Filter.Prefix
				}
				if resf1iter.Filter.Tag != nil {
					resf1elemf3f2 := &svcapitypes.Tag{}
					if resf1iter.Filter.Tag.Key != nil {
						resf1elemf3f2.Key = resf1iter.Filter.Tag.Key
					}
					if resf1iter.Filter.Tag.Value != nil {
						resf1elemf3f2.Value = resf1iter.Filter.Tag.Value
					}
					resf1elemf3.Tag = resf1elemf3f2
				}
				resf1elem.Filter = resf1elemf3
			}
			if resf1iter.ID != nil {
				resf1elem.ID = resf1iter.ID
			}
			if resf1iter.Prefix != nil {
				resf1elem.Prefix = resf1iter.Prefix
			}
			if resf1iter.Priority != nil {
				priorityCopy := int64(*resf1iter.Priority)
				resf1elem.Priority = &priorityCopy
			}
			if resf1iter.SourceSelectionCriteria != nil {
				resf1elemf7 := &svcapitypes.SourceSelectionCriteria{}
				if resf1iter.SourceSelectionCriteria.ReplicaModifications != nil {
					resf1elemf7f0 := &svcapitypes.ReplicaModifications{}
					if resf1iter.SourceSelectionCriteria.ReplicaModifications.Status != "" {
						resf1elemf7f0.Status = aws.String(string(resf1iter.SourceSelectionCriteria.ReplicaModifications.Status))
					}
					resf1elemf7.ReplicaModifications = resf1elemf7f0
				}
				if resf1iter.SourceSelectionCriteria.SseKmsEncryptedObjects != nil {
					resf1elemf7f1 := &svcapitypes.SSEKMSEncryptedObjects{}
					if resf1iter.SourceSelectionCriteria.SseKmsEncryptedObjects.Status != "" {
						resf1elemf7f1.Status = aws.String(string(resf1iter.SourceSelectionCriteria.SseKmsEncryptedObjects.Status))
					}
					resf1elemf7.SSEKMSEncryptedObjects = resf1elemf7f1
				}
				resf1elem.SourceSelectionCriteria = resf1elemf7
			}
			if resf1iter.Status != "" {
				resf1elem.Status = aws.String(string(resf1iter.Status))
			}
			resf1 = append(resf1, resf1elem)
		}
		res.Rules = resf1
	}

	return res
}

// newRequestPaymentConfiguration returns a RequestPaymentConfiguration object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newRequestPaymentConfiguration(
	r *resource,
) *svcsdktypes.RequestPaymentConfiguration {
	res := &svcsdktypes.RequestPaymentConfiguration{}

	if r.ko.Spec.RequestPayment.Payer != nil {
		res.Payer = svcsdktypes.Payer(*r.ko.Spec.RequestPayment.Payer)
	}

	return res
}

// setResourceRequestPayment sets the `RequestPayment` spec field
// given the output of a `GetBucketRequestPayment` operation.
func (rm *resourceManager) setResourceRequestPayment(
	r *resource,
	resp *svcsdk.GetBucketRequestPaymentOutput,
) *svcapitypes.RequestPaymentConfiguration {
	res := &svcapitypes.RequestPaymentConfiguration{}

	if resp.Payer != "" {
		res.Payer = aws.String(string(resp.Payer))
	}

	return res
}

// newTagging returns a Tagging object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newTagging(
	r *resource,
) *svcsdktypes.Tagging {
	res := &svcsdktypes.Tagging{}

	if r.ko.Spec.Tagging.TagSet != nil {
		resf0 := []svcsdktypes.Tag{}
		for _, resf0iter := range r.ko.Spec.Tagging.TagSet {
			resf0elem := &svcsdktypes.Tag{}
			if resf0iter.Key != nil {
				resf0elem.Key = resf0iter.Key
			}
			if resf0iter.Value != nil {
				resf0elem.Value = resf0iter.Value
			}
			resf0 = append(resf0, *resf0elem)
		}
		res.TagSet = resf0
	}

	return res
}

// setResourceTagging sets the `Tagging` spec field
// given the output of a `GetBucketTagging` operation.
func (rm *resourceManager) setResourceTagging(
	r *resource,
	resp *svcsdk.GetBucketTaggingOutput,
) *svcapitypes.Tagging {
	res := &svcapitypes.Tagging{}

	if resp.TagSet != nil {
		resf0 := []*svcapitypes.Tag{}
		for _, resf0iter := range resp.TagSet {
			resf0elem := &svcapitypes.Tag{}
			if resf0iter.Key != nil {
				resf0elem.Key = resf0iter.Key
			}
			if resf0iter.Value != nil {
				resf0elem.Value = resf0iter.Value
			}
			resf0 = append(resf0, resf0elem)
		}
		res.TagSet = resf0
	}

	return res
}

// newVersioningConfiguration returns a VersioningConfiguration object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newVersioningConfiguration(
	r *resource,
) *svcsdktypes.VersioningConfiguration {
	res := &svcsdktypes.VersioningConfiguration{}

	if r.ko.Spec.Versioning.Status != nil {
		res.Status = svcsdktypes.BucketVersioningStatus(*r.ko.Spec.Versioning.Status)
	}

	return res
}

// setResourceVersioning sets the `Versioning` spec field
// given the output of a `GetBucketVersioning` operation.
func (rm *resourceManager) setResourceVersioning(
	r *resource,
	resp *svcsdk.GetBucketVersioningOutput,
) *svcapitypes.VersioningConfiguration {
	res := &svcapitypes.VersioningConfiguration{}

	if resp.Status != "" {
		res.Status = aws.String(string(resp.Status))
	}

	return res
}

// newWebsiteConfiguration returns a WebsiteConfiguration object
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) newWebsiteConfiguration(
	r *resource,
) *svcsdktypes.WebsiteConfiguration {
	res := &svcsdktypes.WebsiteConfiguration{}

	if r.ko.Spec.Website.ErrorDocument != nil {
		resf0 := &svcsdktypes.ErrorDocument{}
		if r.ko.Spec.Website.ErrorDocument.Key != nil {
			resf0.Key = r.ko.Spec.Website.ErrorDocument.Key
		}
		res.ErrorDocument = resf0
	}
	if r.ko.Spec.Website.IndexDocument != nil {
		resf1 := &svcsdktypes.IndexDocument{}
		if r.ko.Spec.Website.IndexDocument.Suffix != nil {
			resf1.Suffix = r.ko.Spec.Website.IndexDocument.Suffix
		}
		res.IndexDocument = resf1
	}
	if r.ko.Spec.Website.RedirectAllRequestsTo != nil {
		resf2 := &svcsdktypes.RedirectAllRequestsTo{}
		if r.ko.Spec.Website.RedirectAllRequestsTo.HostName != nil {
			resf2.HostName = r.ko.Spec.Website.RedirectAllRequestsTo.HostName
		}
		if r.ko.Spec.Website.RedirectAllRequestsTo.Protocol != nil {
			resf2.Protocol = svcsdktypes.Protocol(*r.ko.Spec.Website.RedirectAllRequestsTo.Protocol)
		}
		res.RedirectAllRequestsTo = resf2
	}
	if r.ko.Spec.Website.RoutingRules != nil {
		resf3 := []svcsdktypes.RoutingRule{}
		for _, resf3iter := range r.ko.Spec.Website.RoutingRules {
			resf3elem := &svcsdktypes.RoutingRule{}
			if resf3iter.Condition != nil {
				resf3elemf0 := &svcsdktypes.Condition{}
				if resf3iter.Condition.HTTPErrorCodeReturnedEquals != nil {
					resf3elemf0.HttpErrorCodeReturnedEquals = resf3iter.Condition.HTTPErrorCodeReturnedEquals
				}
				if resf3iter.Condition.KeyPrefixEquals != nil {
					resf3elemf0.KeyPrefixEquals = resf3iter.Condition.KeyPrefixEquals
				}
				resf3elem.Condition = resf3elemf0
			}
			if resf3iter.Redirect != nil {
				resf3elemf1 := &svcsdktypes.Redirect{}
				if resf3iter.Redirect.HostName != nil {
					resf3elemf1.HostName = resf3iter.Redirect.HostName
				}
				if resf3iter.Redirect.HTTPRedirectCode != nil {
					resf3elemf1.HttpRedirectCode = resf3iter.Redirect.HTTPRedirectCode
				}
				if resf3iter.Redirect.Protocol != nil {
					resf3elemf1.Protocol = svcsdktypes.Protocol(*resf3iter.Redirect.Protocol)
				}
				if resf3iter.Redirect.ReplaceKeyPrefixWith != nil {
					resf3elemf1.ReplaceKeyPrefixWith = resf3iter.Redirect.ReplaceKeyPrefixWith
				}
				if resf3iter.Redirect.ReplaceKeyWith != nil {
					resf3elemf1.ReplaceKeyWith = resf3iter.Redirect.ReplaceKeyWith
				}
				resf3elem.Redirect = resf3elemf1
			}
			resf3 = append(resf3, *resf3elem)
		}
		res.RoutingRules = resf3
	}

	return res
}

// setResourceWebsite sets the `Website` spec field
// given the output of a `GetBucketWebsite` operation.
func (rm *resourceManager) setResourceWebsite(
	r *resource,
	resp *svcsdk.GetBucketWebsiteOutput,
) *svcapitypes.WebsiteConfiguration {
	res := &svcapitypes.WebsiteConfiguration{}

	if resp.ErrorDocument != nil {
		resf0 := &svcapitypes.ErrorDocument{}
		if resp.ErrorDocument.Key != nil {
			resf0.Key = resp.ErrorDocument.Key
		}
		res.ErrorDocument = resf0
	}
	if resp.IndexDocument != nil {
		resf1 := &svcapitypes.IndexDocument{}
		if resp.IndexDocument.Suffix != nil {
			resf1.Suffix = resp.IndexDocument.Suffix
		}
		res.IndexDocument = resf1
	}
	if resp.RedirectAllRequestsTo != nil {
		resf2 := &svcapitypes.RedirectAllRequestsTo{}
		if resp.RedirectAllRequestsTo.HostName != nil {
			resf2.HostName = resp.RedirectAllRequestsTo.HostName
		}
		if resp.RedirectAllRequestsTo.Protocol != "" {
			resf2.Protocol = aws.String(string(resp.RedirectAllRequestsTo.Protocol))
		}
		res.RedirectAllRequestsTo = resf2
	}
	if resp.RoutingRules != nil {
		resf3 := []*svcapitypes.RoutingRule{}
		for _, resf3iter := range resp.RoutingRules {
			resf3elem := &svcapitypes.RoutingRule{}
			if resf3iter.Condition != nil {
				resf3elemf0 := &svcapitypes.Condition{}
				if resf3iter.Condition.HttpErrorCodeReturnedEquals != nil {
					resf3elemf0.HTTPErrorCodeReturnedEquals = resf3iter.Condition.HttpErrorCodeReturnedEquals
				}
				if resf3iter.Condition.KeyPrefixEquals != nil {
					resf3elemf0.KeyPrefixEquals = resf3iter.Condition.KeyPrefixEquals
				}
				resf3elem.Condition = resf3elemf0
			}
			if resf3iter.Redirect != nil {
				resf3elemf1 := &svcapitypes.Redirect{}
				if resf3iter.Redirect.HostName != nil {
					resf3elemf1.HostName = resf3iter.Redirect.HostName
				}
				if resf3iter.Redirect.HttpRedirectCode != nil {
					resf3elemf1.HTTPRedirectCode = resf3iter.Redirect.HttpRedirectCode
				}
				if resf3iter.Redirect.Protocol != "" {
					resf3elemf1.Protocol = aws.String(string(resf3iter.Redirect.Protocol))
				}
				if resf3iter.Redirect.ReplaceKeyPrefixWith != nil {
					resf3elemf1.ReplaceKeyPrefixWith = resf3iter.Redirect.ReplaceKeyPrefixWith
				}
				if resf3iter.Redirect.ReplaceKeyWith != nil {
					resf3elemf1.ReplaceKeyWith = resf3iter.Redirect.ReplaceKeyWith
				}
				resf3elem.Redirect = resf3elemf1
			}
			resf3 = append(resf3, resf3elem)
		}
		res.RoutingRules = resf3
	}

	return res
}
