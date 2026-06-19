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
Common chart identity labels.
*/}}
{{- define "lucity-app.labels" -}}
helm.sh/chart: {{ include "lucity-app.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: {{ include "lucity-app.name" . }}
{{- end }}

{{/*
Selector labels for a specific component.
*/}}
{{- define "lucity-app.selectorLabels" -}}
app.kubernetes.io/name: {{ .name }}
app.kubernetes.io/instance: {{ .release }}
{{- end }}

{{/*
Container image reference for a service.

Pins by digest whenever one is known: Kubernetes pulls by digest while the
tag stays visible for humans ("repo:tag@sha256:..."). Before the first build
has reported a digest the bare tag is used, and a service still awaiting its
initial build (no tag, no digest) renders just the repository. That service is
scaled to zero, so the implicit ":latest" is never pulled.
*/}}
{{- define "lucity-app.image" -}}
{{- $image := . -}}
{{- if $image.digest -}}
{{- if $image.tag -}}
{{- printf "%s:%s@%s" $image.repository $image.tag $image.digest -}}
{{- else -}}
{{- printf "%s@%s" $image.repository $image.digest -}}
{{- end -}}
{{- else if $image.tag -}}
{{- printf "%s:%s" $image.repository $image.tag -}}
{{- else -}}
{{- $image.repository -}}
{{- end -}}
{{- end }}
