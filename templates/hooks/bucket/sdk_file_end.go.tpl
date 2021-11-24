{{ $CRD := .CRD }}
{{ $SDKAPI := .SDKAPI }}

{{ range $specFieldName, $specField := $CRD.Config.Resources.Bucket.Fields -}}

{{/* If the field comes from a single Put* operation */}}
{{- if $specField.From }}
{{- $operationName := $specField.From.Operation }}
{{- $path := $specField.From.Path }}
{{/* Only generate for Put* operation fields */}}
{{- if (eq (slice $operationName 0 3) "Put") }}
{{- $field := (index $CRD.SpecFields $specFieldName )}}
{{- $operation := (index $SDKAPI.API.Operations $operationName) -}}

{{/* Find the structure field within the operation */}}
{{- range $memberRefName, $memberRef := $operation.InputRef.Shape.MemberRefs -}}
{{- if (eq $memberRef.Shape.Type "structure") }}

// new{{ $memberRefName }} returns a {{ $memberRefName }} object 
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) new{{ $memberRefName }}(
    r *resource,
) *svcsdk.{{ $memberRef.ShapeName }} {
    res := &svcsdk.{{ $memberRef.ShapeName }}{}

{{ GoCodeSetSDKForStruct $CRD "" "res" $memberRef "" (printf "r.ko.Spec.%s" $specFieldName) 1 }}

    return res
}

{{/* Find the matching Get* operation */}}
{{- $describeOperationName := (printf "Get%s" (slice $operationName 3))}}
{{- $field := (index $CRD.SpecFields $specFieldName )}}
{{- $operation := (index $SDKAPI.API.Operations $describeOperationName)}}

// setResource{{ $specFieldName }} sets the `{{ $specFieldName }}` spec field
// given the output of a `{{ $operation.Name }}` operation.
func (rm *resourceManager) setResource{{ $specFieldName }}(
    r *resource,
    resp *svcsdk.{{ $operation.OutputRef.ShapeName }},
) *svcapitypes.{{ $memberRef.ShapeName }} {
    res := &svcapitypes.{{ $memberRef.ShapeName }}{}

{{/* Some operations have wrapping structures in their response */}}
{{- if (eq $operationName "PutBucketEncryption") }}
{{ GoCodeSetResourceForStruct $CRD $specFieldName "res" $memberRef "resp.ServerSideEncryptionConfiguration" $memberRef 1 }}
{{- else if (eq $operationName "PutBucketOwnershipControls") }}
{{ GoCodeSetResourceForStruct $CRD $specFieldName "res" $memberRef "resp.OwnershipControls" $memberRef 1 }}
{{- else if (eq $operationName "PutBucketReplication") }}
{{ GoCodeSetResourceForStruct $CRD $specFieldName "res" $memberRef "resp.ReplicationConfiguration" $memberRef 1 }}
{{- else if (eq $operationName "PutPublicAccessBlock") }}
{{ GoCodeSetResourceForStruct $CRD $specFieldName "res" $memberRef "resp.PublicAccessBlockConfiguration" $memberRef 1 }}
{{- else }}
{{ GoCodeSetResourceForStruct $CRD $specFieldName "res" $memberRef "resp" $memberRef 1 }}
{{ end }}

    return res
}
{{- end }}
{{- end }}
{{- end }}

{{/* If the field is a custom shape */}}
{{- else if $specField.CustomField }}

{{- $memberRefName := $specField.CustomField.ListOf }}
{{/* Iterate through the custom shapes to find the matching shape ref */}}
{{- range $index, $customShape := $SDKAPI.CustomShapes }}
{{- if (eq (Dereference $customShape.MemberShapeName) $memberRefName) }}

{{- $memberRef := $customShape.Shape.MemberRef }}

// new{{ $memberRefName }} returns a {{ $memberRefName }} object 
// with each the field set by the corresponding configuration's fields.
func (rm *resourceManager) new{{ $memberRefName }}(
    c *svcapitypes.{{ $memberRefName }},
) *svcsdk.{{ $memberRefName }} {
    res := &svcsdk.{{ $memberRefName }}{}

{{ GoCodeSetSDKForStruct $CRD "" "res" $memberRef "" "c" 1 }}

    return res
}

// set{{ $memberRefName }} sets a resource {{ $memberRefName }} type
// given the SDK type.
func (rm *resourceManager) setResource{{ $memberRefName }}(
    r *resource,
    resp *svcsdk.{{ $memberRefName }},
) *svcapitypes.{{ $memberRefName }} {
    res := &svcapitypes.{{ $memberRefName }}{}

{{ GoCodeSetResourceForStruct $CRD $specFieldName "res" $memberRef "resp" $memberRef 1 }}

    return res
}

func compare{{$memberRefName}} (
	a *svcapitypes.{{ $memberRefName }},
	b *svcapitypes.{{ $memberRefName }},
) *ackcompare.Delta {
	delta := ackcompare.NewDelta()
{{ GoCodeCompareStruct $CRD $memberRef.Shape "delta" "a" "b" $memberRefName 1 }}
	return delta
}

// get{{$memberRefName}}Action returns the determined action for a given
// configuration object, depending on the desired and latest values
func get{{$memberRefName}}Action(
    c *svcapitypes.{{ $memberRefName }},
    latest *resource,
) ConfigurationAction{
    action := ConfigurationActionPut
	if latest != nil {
		for _, l := range latest.ko.Spec.{{ $specFieldName }} {
			if *l.ID != *c.ID {
				continue
			}

			// Don't perform any action if they are identical
			delta := compare{{$memberRefName}}(l, c)
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

func (rm *resourceManager) newListBucket{{ $specFieldName }}Payload(
	r *resource,
) *svcsdk.ListBucket{{ $memberRefName }}sInput {
	res := &svcsdk.ListBucket{{ $memberRefName }}sInput{}
	res.SetBucket(*r.ko.Spec.Name)
	return res
}

func (rm *resourceManager) newPutBucket{{ $specFieldName }}Payload(
	r *resource,
	c svcapitypes.{{ $memberRefName }},
) *svcsdk.PutBucket{{ $memberRefName }}Input {
	res := &svcsdk.PutBucket{{ $memberRefName }}Input{}
	res.SetBucket(*r.ko.Spec.Name)
	res.SetId(*c.ID)
	res.Set{{ $memberRefName }}(rm.new{{ $memberRefName }}(&c))

	return res
}

func (rm *resourceManager) newDeleteBucket{{ $specFieldName }}Payload(
	r *resource,
	c svcapitypes.{{ $memberRefName }},
) *svcsdk.DeleteBucket{{ $memberRefName }}Input {
	res := &svcsdk.DeleteBucket{{ $memberRefName }}Input{}
	res.SetBucket(*r.ko.Spec.Name)
	res.SetId(*c.ID)

	return res
}

func (rm *resourceManager) delete{{ $memberRefName }}(
	ctx context.Context,
	r *resource,
	c svcapitypes.{{ $memberRefName }},
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.delete{{ $memberRefName }}")
	defer exit(err)

	input := rm.newDeleteBucket{{ $specFieldName }}Payload(r, c)
	_, err = rm.sdkapi.DeleteBucket{{ $memberRefName }}WithContext(ctx, input)
	rm.metrics.RecordAPICall("DELETE", "DeleteBucket{{ $memberRefName }}", err)
	return err
}

func (rm *resourceManager) put{{ $memberRefName }}(
	ctx context.Context,
	r *resource,
	c svcapitypes.{{ $memberRefName }},
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.put{{ $memberRefName }}")
	defer exit(err)

	input := rm.newPutBucket{{ $specFieldName }}Payload(r, c)
	_, err = rm.sdkapi.PutBucket{{ $memberRefName }}WithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "PutBucket{{ $memberRefName }}", err)
	return err
}

func (rm *resourceManager) sync{{ $specFieldName }}(
	ctx context.Context,
	desired *resource,
	latest *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.sync{{ $specFieldName }}")
	defer exit(err)

	for _, c := range desired.ko.Spec.{{ $specFieldName }} {
		action := get{{ $memberRefName }}Action(c, latest)

		switch action {
		case ConfigurationActionUpdate:
			fallthrough
		case ConfigurationActionPut:
			if err = rm.put{{ $memberRefName }}(ctx, desired, *c); err != nil {
				return err
			}
		default:
		}
	}

	if latest != nil {
		// Find any configurations that are in the latest but not in desired
		for _, l := range latest.ko.Spec.{{ $specFieldName }} {
			exists := false
			for _, c := range desired.ko.Spec.{{ $specFieldName }} {
				if *c.ID != *l.ID {
					continue
				}
				exists = true
				break
			}

			if !exists {
				if err = rm.delete{{ $memberRefName }}(ctx, desired, *l); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

{{- end }}
{{- end }}

{{- end }}

{{- end }}