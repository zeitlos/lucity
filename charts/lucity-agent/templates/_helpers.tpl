{{- define "lucity-agent.fullname" -}}
{{- default .Release.Name .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "lucity-agent.clusterName" -}}
{{- printf "%s-%s" .Release.Namespace (include "lucity-agent.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "lucity-agent.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "lucity-agent.selectorLabels" -}}
app.kubernetes.io/name: {{ default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "lucity-agent.labels" -}}
helm.sh/chart: {{ include "lucity-agent.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/part-of: lucity
app.kubernetes.io/component: agent
{{ include "lucity-agent.selectorLabels" . }}
{{- end }}
