{{- define "lucity-infra.otel.nodeLocalPods" -}}
kubernetes_sd_configs:
  - role: pod
    namespaces:
      names: [{{ . }}]
    selectors:
      - role: pod
        field: spec.nodeName=${env:K8S_NODE_NAME}
{{- end }}

{{- define "lucity-infra.otel.selfTelemetry" -}}
metrics:
  level: normal
  readers:
    - periodic:
        interval: 30000
        exporter:
          otlp:
            protocol: http/protobuf
            endpoint: http://{{ .Release.Name }}-victoria-metrics-single-server:8428/opentelemetry/v1/metrics
{{- end }}
