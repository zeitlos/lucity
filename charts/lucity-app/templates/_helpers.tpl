{{/*
Expand the name of the chart.
*/}}
{{- define "lucity-app.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "lucity-app.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "lucity-app.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels. Reads workspace/project/environment from top-level values
written by the packager. Used by the platform package for discovery.
*/}}
{{- define "lucity-app.labels" -}}
helm.sh/chart: {{ include "lucity-app.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: {{ include "lucity-app.name" . }}
{{- if .Values.workspace }}
lucity.dev/workspace: {{ .Values.workspace | quote }}
{{- end }}
{{- if .Values.project }}
lucity.dev/project: {{ .Values.project | quote }}
{{- end }}
{{- if .Values.environment }}
lucity.dev/environment: {{ .Values.environment | quote }}
{{- end }}
{{- end }}

{{/*
Selector labels for a specific component.
*/}}
{{- define "lucity-app.selectorLabels" -}}
app.kubernetes.io/name: {{ .name }}
app.kubernetes.io/instance: {{ .release }}
{{- end }}
