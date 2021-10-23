package scaffold

import (
	"path/filepath"
)

const (
	defaultNotes     = `Your {{ .Release.Name }} installation`
	sep              = string(filepath.Separator)
	defaultChartfile = `apiVersion: v1
name: {{ .Values.name }}
description: A Helm chart for Kubernetes
version: 0.1.0
appVersion: 1.16.0
dependencies:
{{ range $_ind, $dep := .Values.dependencies }}
- name: {{ $dep.Name }}
  repository: {{ $dep.Repository }}
  version: {{ $dep.Version }}
{{ end }}
`
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
