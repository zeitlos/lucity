{{- define "lucity-workspace.workspace" -}}
{{- required "workspace is required" .Values.workspace -}}
{{- end }}

{{- define "lucity-workspace.labels" -}}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: {{ .Chart.Name }}
lucity.dev/workspace: {{ include "lucity-workspace.workspace" . }}
{{- end }}

{{- define "lucity-workspace.port" -}}
{{- if eq .host .hostname -}}
443
{{- else -}}
{{- trimPrefix (printf "%s:" .hostname) .host -}}
{{- end -}}
{{- end }}

{{- define "lucity-workspace.route" -}}
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: {{ .resource }}
  labels:
    {{- include "lucity-workspace.labels" .root | nindent 4 }}
spec:
  parentRefs:
    - name: {{ .root.Values.gateway.name }}
      sectionName: {{ .root.Values.gateway.listener }}
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: {{ .prefix }}
      filters:
        - type: URLRewrite
          urlRewrite:
            path:
              type: ReplacePrefixMatch
              replacePrefixMatch: /
      backendRefs:
        - group: agentgateway.dev
          kind: AgentgatewayBackend
          name: {{ .resource }}
{{- end }}
