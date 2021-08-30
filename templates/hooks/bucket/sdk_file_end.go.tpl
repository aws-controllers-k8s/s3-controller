{{ $CRD := .CRD }}
{{ $SDKAPI := .SDKAPI }}

{{ range $specFieldName, $specField := $CRD.Config.Resources.Bucket.Fields -}}
{{- $operationName := $specField.From.Operation }}
{{- $path := $specField.From.Path }}
{{- if (eq (slice $operationName 0 3) "Put") }}
{{- $field := (index $CRD.SpecFields $specFieldName )}}
{{- $operation := (index $SDKAPI.API.Operations $operationName) -}}

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

{{- if (eq $operationName "PutBucketEncryption") }}
{{ GoCodeSetResourceForStruct $CRD "" "res" $memberRef "resp.ServerSideEncryptionConfiguration" $memberRef 1 }}
{{- else if (eq $operationName "PutBucketOwnershipControls") }}
{{ GoCodeSetResourceForStruct $CRD "" "res" $memberRef "resp.OwnershipControls" $memberRef 1 }}
{{- else if (eq $operationName "PutBucketReplication") }}
{{ GoCodeSetResourceForStruct $CRD "" "res" $memberRef "resp.ReplicationConfiguration" $memberRef 1 }}
{{- else }}
{{ GoCodeSetResourceForStruct $CRD "" "res" $memberRef "resp" $memberRef 1 }}
{{ end }}

    return res
}
{{- end }}
{{- end }}
{{- end }}
{{- end }}