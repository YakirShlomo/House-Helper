{{/*
Expand the name of the chart.
*/}}
{{- define "house-helper.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "house-helper.fullname" -}}
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
{{- define "house-helper.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "house-helper.labels" -}}
helm.sh/chart: {{ include "house-helper.chart" . }}
{{ include "house-helper.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: house-helper
environment: {{ .Values.environment }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "house-helper.selectorLabels" -}}
app.kubernetes.io/name: {{ include "house-helper.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Component labels
*/}}
{{- define "house-helper.componentLabels" -}}
{{- $component := . }}
app.kubernetes.io/component: {{ $component }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "house-helper.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "house-helper.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Namespace
*/}}
{{- define "house-helper.namespace" -}}
{{- default .Release.Namespace .Values.namespaceOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Image name
*/}}
{{- define "house-helper.image" -}}
{{- $registry := .registry }}
{{- $repository := .repository }}
{{- $tag := .tag }}
{{- if $registry }}
{{- printf "%s/%s:%s" $registry $repository $tag }}
{{- else }}
{{- printf "%s:%s" $repository $tag }}
{{- end }}
{{- end }}

{{/*
Environment variables from secrets
*/}}
{{- define "house-helper.envFromSecrets" -}}
- secretRef:
    name: app-secrets
- secretRef:
    name: db-credentials
- secretRef:
    name: redis-credentials
- secretRef:
    name: kafka-credentials
{{- end }}

{{/*
Common environment variables
*/}}
{{- define "house-helper.commonEnv" -}}
- name: ENVIRONMENT
  value: {{ .Values.environment | quote }}
- name: PROJECT_NAME
  value: {{ .Values.global.projectName | quote }}
- name: POD_NAME
  valueFrom:
    fieldRef:
      fieldPath: metadata.name
- name: POD_NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
- name: POD_IP
  valueFrom:
    fieldRef:
      fieldPath: status.podIP
- name: NODE_NAME
  valueFrom:
    fieldRef:
      fieldPath: spec.nodeName
{{- end }}

{{/*
Database connection string
*/}}
{{- define "house-helper.dbConnectionString" -}}
{{- if .Values.postgresql.enabled }}
postgresql://$(DB_USER):$(DB_PASSWORD)@{{ .Release.Name }}-postgresql:5432/$(DB_NAME)?sslmode=disable
{{- else }}
postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=require
{{- end }}
{{- end }}

{{/*
Redis connection string
*/}}
{{- define "house-helper.redisConnectionString" -}}
{{- if .Values.redis.enabled }}
redis://:$(REDIS_PASSWORD)@{{ .Release.Name }}-redis-master:6379/0
{{- else }}
redis://:$(REDIS_AUTH_TOKEN)@$(REDIS_ENDPOINT):$(REDIS_PORT)/0
{{- end }}
{{- end }}

{{/*
Kafka bootstrap servers
*/}}
{{- define "house-helper.kafkaBootstrapServers" -}}
{{- if .Values.kafka.enabled }}
{{ .Release.Name }}-kafka:9092
{{- else }}
$(KAFKA_BOOTSTRAP_SERVERS)
{{- end }}
{{- end }}

{{/*
Resource limits and requests
*/}}
{{- define "house-helper.resources" -}}
{{- if .resources }}
resources:
  {{- if .resources.requests }}
  requests:
    {{- if .resources.requests.cpu }}
    cpu: {{ .resources.requests.cpu }}
    {{- end }}
    {{- if .resources.requests.memory }}
    memory: {{ .resources.requests.memory }}
    {{- end }}
  {{- end }}
  {{- if .resources.limits }}
  limits:
    {{- if .resources.limits.cpu }}
    cpu: {{ .resources.limits.cpu }}
    {{- end }}
    {{- if .resources.limits.memory }}
    memory: {{ .resources.limits.memory }}
    {{- end }}
  {{- end }}
{{- end }}
{{- end }}

{{/*
Pod annotations
*/}}
{{- define "house-helper.podAnnotations" -}}
prometheus.io/scrape: "true"
prometheus.io/port: "{{ .port }}"
prometheus.io/path: "/metrics"
checksum/config: {{ include (print $.Template.BasePath "/configmap.yaml") . | sha256sum }}
{{- end }}

{{/*
Security context
*/}}
{{- define "house-helper.securityContext" -}}
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  runAsGroup: 3000
  fsGroup: 2000
  capabilities:
    drop:
      - ALL
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
{{- end }}

{{/*
Pod security context
*/}}
{{- define "house-helper.podSecurityContext" -}}
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 2000
  seccompProfile:
    type: RuntimeDefault
{{- end }}
