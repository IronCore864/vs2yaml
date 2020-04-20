apiVersion: v1
kind: Secret
metadata:
  name: {{ .name }}
  namespace: {{ .namespace }}
type: Opaque
data:
{{ range  $k, $v := .data }}  {{ $k }}: {{ $v }}
{{ end }}
