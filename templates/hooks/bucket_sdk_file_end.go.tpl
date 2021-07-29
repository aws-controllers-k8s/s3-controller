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

// create{{ $memberRefName }} returns a {{ $memberRefName }} object 
// with each the field set by the resource's corresponding spec field.
func (rm *resourceManager) create{{ $memberRefName }}(
    r *resource,
) *svcsdk.{{ $memberRef.ShapeName }} {
    res := &svcsdk.{{ $memberRef.ShapeName }}{}

{{ GoCodeSetOperationStruct $CRD "" "res" $memberRef "" (printf "r.ko.Spec.%s" $specFieldName) 1}}
    return res
}
{{- end}}
{{- end}}
{{- end }}
{{- end }}