package scaffold

import (
	"path/filepath"
)

const (
	defaultNotes = `Your {{ .Release.Name }} installation`
	sep          = string(filepath.Separator)

	defaultIgnore = `# Patterns to ignore when building packages.
# This supports shell glob matching, relative path matching, and
# negation (prefixed with !). Only one pattern per line.
.DS_Store
# Common VCS dirs
.git/
.gitignore
.bzr/
.bzrignore
.hg/
.hgignore
.svn/
# Common backup files
*.swp
*.bak
*.tmp
*~
# Various IDEs
.project
.idea/
*.tmproj
.vscode/
`
	defaultApplication = `apiVersion: app.k8s.io/v1beta1
kind: Application
metadata:
  name: {{ .Name }}
spec:
  selector:
    matchLabels: {}
  componentKinds:
  - group: v1
    kind: Service
  - group: networking.k8s.io
    kind: Ingress
  - group: cert-manager.io
    kind: Certificate
  - group: apps
    kind: StatefulSet
  - group: apps
    kind: Deployment
  - group: batch
    kind: CronJob
  - group: batch
    kind: Job
  descriptor:
    type: {{ .Name }}
    version: {{ .Version }}
    description: {{ .Description }}
    icons:
    - src: {{ .Icon }}
    {{ if .DarkIcon }}
    - src: {{ .DarkIcon }}
    {{ end }}
`

	appTemplate = `
    {{- if .Values.global }}
    {{- if .Values.global.application }}
    {{- if .Values.global.application.links }}
    links:
    {{ toYaml .Values.global.application.links | nindent 6 }}
    {{- end }}
  {{- if .Values.global.application.info }}
  info:
  {{ toYaml .Values.global.application.info | nindent 4 }}
  {{- end }}
  {{- end }}
  {{- end }}
`

	licenseSecret = `apiVersion: v1
kind: Secret
metadata:
  name: plural-license-secret
stringData:
  license: {{ .Values.plrl.license }}
`

	license = `apiVersion: platform.plural.sh/v1alpha1
kind: License
metadata:
  name: %s
spec:
  secretRef:
    name: plural-license-secret
    key: license
`

	// ChartfileName is the default Chart file name.
	ChartfileName = "Chart.yaml"
	// ValuesfileName is the default values file name.
	TemplatesDir = "templates"
	// ChartsDir is the relative directory name for charts dependencies.
	ChartsDir = "charts"
	// IgnorefileName is the name of the Helm ignore file.
	IgnorefileName = ".helmignore"
	// NotesName is the name of the example NOTES.txt file.
	NotesName = TemplatesDir + sep + "NOTES.txt"
	// file to put the default application resource in
	ApplicationName = TemplatesDir + sep + "application.yaml"
	// file to put the license secret in
	LicenseSecretName = TemplatesDir + sep + "secret.yaml"
	// file to put the license crd in
	LicenseCrdName = TemplatesDir + sep + "license.yaml"
)
