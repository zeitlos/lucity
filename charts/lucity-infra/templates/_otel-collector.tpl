{{- define "lucity-infra.otel.nodeLocalPods" -}}
kubernetes_sd_configs:
  - role: pod
    namespaces:
      names: [{{ . }}]
    selectors:
      - role: pod
        field: spec.nodeName=${env:K8S_NODE_NAME}
{{- end }}
