{{/*
Expand the name of the chart.
*/}}
{{- define "ollamacron.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "ollamacron.fullname" -}}
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
{{- define "ollamacron.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "ollamacron.labels" -}}
helm.sh/chart: {{ include "ollamacron.chart" . }}
{{ include "ollamacron.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "ollamacron.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ollamacron.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "ollamacron.serviceAccountName" -}}
{{- if .Values.ollamacron.serviceAccount.create }}
{{- default (include "ollamacron.fullname" .) .Values.ollamacron.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.ollamacron.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the image name
*/}}
{{- define "ollamacron.image" -}}
{{- $registry := .Values.global.imageRegistry | default .Values.ollamacron.image.registry -}}
{{- $repository := .Values.ollamacron.image.repository -}}
{{- $tag := .Values.ollamacron.image.tag | default .Chart.AppVersion -}}
{{- if $registry }}
{{- printf "%s/%s:%s" $registry $repository $tag }}
{{- else }}
{{- printf "%s:%s" $repository $tag }}
{{- end }}
{{- end }}

{{/*
Create the storage class name
*/}}
{{- define "ollamacron.storageClass" -}}
{{- if .Values.ollamacron.persistence.storageClass }}
{{- if (eq "-" .Values.ollamacron.persistence.storageClass) }}
{{- printf "storageClassName: \"\"" }}
{{- else }}
{{- printf "storageClassName: %s" .Values.ollamacron.persistence.storageClass }}
{{- end }}
{{- else if .Values.global.storageClass }}
{{- printf "storageClassName: %s" .Values.global.storageClass }}
{{- end }}
{{- end }}

{{/*
Create prometheus rule namespace
*/}}
{{- define "ollamacron.prometheusRule.namespace" -}}
{{- if .Values.prometheusRule.namespace }}
{{- .Values.prometheusRule.namespace }}
{{- else }}
{{- .Release.Namespace }}
{{- end }}
{{- end }}

{{/*
Create service monitor namespace
*/}}
{{- define "ollamacron.serviceMonitor.namespace" -}}
{{- if .Values.serviceMonitor.namespace }}
{{- .Values.serviceMonitor.namespace }}
{{- else }}
{{- .Release.Namespace }}
{{- end }}
{{- end }}