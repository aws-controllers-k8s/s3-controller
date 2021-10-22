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
{{ GoCodeSetResourceForStruct $CRD "" "res" $memberRef "resp.ServerSideEncryptionConfiguration" $memberRef 1 }}
{{- else if (eq $operationName "PutBucketOwnershipControls") }}
{{ GoCodeSetResourceForStruct $CRD "" "res" $memberRef "resp.OwnershipControls" $memberRef 1 }}
{{- else if (eq $operationName "PutBucketReplication") }}
{{ GoCodeSetResourceForStruct $CRD "" "res" $memberRef "resp.ReplicationConfiguration" $memberRef 1 }}
{{- else if (eq $operationName "PutPublicAccessBlock") }}
{{ GoCodeSetResourceForStruct $CRD "" "res" $memberRef "resp.PublicAccessBlockConfiguration" $memberRef 1 }}
{{- else }}
{{ GoCodeSetResourceForStruct $CRD "" "res" $memberRef "resp" $memberRef 1 }}
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

{{ GoCodeSetResourceForStruct $CRD "" "res" $memberRef "resp" $memberRef 1 }}

    return res
}

{{- end }}
{{- end }}

{{- end }}

{{- end }}