{{/*
Controller image pull secrets
*/}}
{{- define "gwin.controller.image.pullSecrets" -}}
  {{- if .Values.controller.image.pullSecrets -}}
    {{- toYaml .Values.controller.image.pullSecrets -}}
  {{- else -}}
    {{- toYaml .Values.global.image.pullSecrets -}}
  {{- end -}}
{{- end }}

{{/*
Controller image pull policy
*/}}
{{- define "gwin.controller.image.pullPolicy" -}}
  {{- if .Values.controller.image.pullPolicy -}}
    {{- .Values.controller.image.pullPolicy -}}
  {{- else -}}
    {{- .Values.global.image.pullPolicy -}}
  {{- end -}}
{{- end }}

{{/*
CRDs auto upgrade image pull secrets
*/}}
{{- define "gwin.crdsAutoUpgrade.image.pullSecrets" -}}
  {{- if .Values.crdsAutoUpgrade.image.pullSecrets -}}
    {{- toYaml .Values.crdsAutoUpgrade.image.pullSecrets -}}
  {{- else -}}
    {{- toYaml .Values.global.image.pullSecrets -}}
  {{- end -}}
{{- end }}

{{/*
CRDs auto upgrade image pull policy
*/}}
{{- define "gwin.crdsAutoUpgrade.image.pullPolicy" -}}
  {{- if .Values.crdsAutoUpgrade.image.pullPolicy -}}
    {{- .Values.crdsAutoUpgrade.image.pullPolicy -}}
  {{- else -}}
    {{- .Values.global.image.pullPolicy -}}
  {{- end -}}
{{- end }}

{{/*
Controller pod security context
*/}}
{{- define "gwin.controller.podSecurityContext" -}}
  {{- if .Values.controller.podSecurityContext -}}
    {{- toYaml .Values.controller.podSecurityContext -}}
  {{- else -}}
    {{- toYaml .Values.global.podSecurityContext -}}
  {{- end -}}
{{- end }}

{{/*
Controller container security context
*/}}
{{- define "gwin.controller.containerSecurityContext" -}}
  {{- if .Values.controller.containerSecurityContext -}}
    {{- toYaml .Values.controller.containerSecurityContext -}}
  {{- else -}}
    {{- toYaml .Values.global.containerSecurityContext -}}
  {{- end -}}
{{- end }}

{{/*
CRDs auto upgrade pod security context
*/}}
{{- define "gwin.crdsAutoUpgrade.podSecurityContext" -}}
  {{- if .Values.crdsAutoUpgrade.podSecurityContext -}}
    {{- toYaml .Values.crdsAutoUpgrade.podSecurityContext -}}
  {{- else -}}
    {{- toYaml .Values.global.podSecurityContext -}}
  {{- end -}}
{{- end }}

{{/*
CRDs auto upgrade container security context
*/}}
{{- define "gwin.crdsAutoUpgrade.containerSecurityContext" -}}
  {{- if .Values.crdsAutoUpgrade.containerSecurityContext -}}
    {{- toYaml .Values.crdsAutoUpgrade.containerSecurityContext -}}
  {{- else -}}
    {{- toYaml .Values.global.containerSecurityContext -}}
  {{- end -}}
{{- end }}

{{/*
Nodecheck image pull secrets
*/}}
{{- define "gwin.nodecheck.image.pullSecrets" -}}
  {{- if .Values.nodecheck.image.pullSecrets -}}
    {{- toYaml .Values.nodecheck.image.pullSecrets -}}
  {{- else -}}
    {{- toYaml .Values.global.image.pullSecrets -}}
  {{- end -}}
{{- end }}

{{/*
Nodecheck image pull policy
*/}}
{{- define "gwin.nodecheck.image.pullPolicy" -}}
  {{- if .Values.nodecheck.image.pullPolicy -}}
    {{- .Values.nodecheck.image.pullPolicy -}}
  {{- else -}}
    {{- .Values.global.image.pullPolicy -}}
  {{- end -}}
{{- end }}

{{/*
Nodecheck pod security context
*/}}
{{- define "gwin.nodecheck.podSecurityContext" -}}
  {{- if .Values.nodecheck.podSecurityContext -}}
    {{- toYaml .Values.nodecheck.podSecurityContext -}}
  {{- else -}}
    {{- toYaml .Values.global.podSecurityContext -}}
  {{- end -}}
{{- end }}

{{/*
Nodecheck container security context
*/}}
{{- define "gwin.nodecheck.containerSecurityContext" -}}
  {{- if .Values.nodecheck.containerSecurityContext -}}
    {{- toYaml .Values.nodecheck.containerSecurityContext -}}
  {{- else -}}
    {{- toYaml .Values.global.containerSecurityContext -}}
  {{- end -}}
{{- end }}

{{/*
Calculate authentication method for Yandex Cloud service account
Logic:
1. If workloadIdentityFederation.serviceAccountID is set -> use "workloadIdentityFederation"
2. If secret.value is set -> use "secret"
3. If secret.create is false -> use "secret" (user guarantees they brought their own secret)
4. Otherwise -> fail with error message
*/}}
{{- define "gwin.controller.authMethod" -}}
  {{- $ycServiceAccount := .Values.controller.ycServiceAccount -}}
  {{- if $ycServiceAccount.workloadIdentityFederation.serviceAccountID -}}
    {{- "workloadIdentityFederation" -}}
  {{- else -}}
    {{- "secret" -}}
  {{- end -}}
{{- end }}
