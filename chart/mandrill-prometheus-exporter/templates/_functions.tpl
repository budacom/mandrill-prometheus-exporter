
{{/* Build the Spotahome standard labels */}}
{{- define "selector-labels" -}}
app.kubernetes.io/name: {{ .Chart.Name | quote }}
app.kubernetes.io/component: {{ .Values.role | quote }}
app.kubernetes.io/instance: {{ .Release.Name | quote }}
{{- end }}

{{- define "common-labels" -}}
team: {{ .Values.team | quote }}
{{- end }}

{{- define "helm-labels" -}}
{{ include "common-labels" . }}
app.kubernetes.io/managed-by: {{ .Release.Service | quote }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | quote }}
{{- end }}

{{/* Build wide-used variables the application */}}
{{- define "name" -}}
{{- if contains .Chart.Name .Release.Name -}}
{{- .Release.Name -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name .Chart.Name -}}
{{- end -}}
{{- end -}}

{{ define "image" -}}
{{ printf "%s:%s" .Values.image .Values.tag }}
{{- end }}
