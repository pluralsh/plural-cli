package main_test

import (
	"os"
	"path/filepath"
	"testing"

	plural "github.com/pluralsh/plural/cmd/plural"
	"github.com/stretchr/testify/assert"
)

const chart_file = `apiVersion: v2
name: minio
description: A helm chart for minio
version: 0.1.0
appVersion: 0.1.53
dependencies:
- name: minio
  version: 0.1.53
  repository: cm://app.plural.sh/cm/minio
  condition: minio.enabled
`

func TestImageBump(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		values        string
		expectedChart string
	}{
		{
			name: `test "image-bump" when new version is higher than current`,
			args: []string{plural.ApplicationName, "utils", "image-bump", "./", "--path", "minio.version", "--tag", "0.1.54"},
			values: `minio:
  enabled: true
  version: 0.1.53`,
			expectedChart: `apiVersion: v2
appVersion: 0.1.53
dependencies:
- condition: minio.enabled
  name: minio
  repository: cm://app.plural.sh/cm/minio
  version: 0.1.53
description: A helm chart for minio
name: minio
version: 0.1.1
`,
		},
		{
			name: `test "image-bump" when new version is equal current`,
			args: []string{plural.ApplicationName, "utils", "image-bump", "./", "--path", "minio.version", "--tag", "0.1.53"},
			values: `minio:
  enabled: true
  version: 0.1.53`,
			expectedChart: chart_file,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			currentDir, err := os.Getwd()
			assert.NoError(t, err)
			dir, err := os.MkdirTemp("", "config")
			assert.NoError(t, err)
			defer func(path, currentDir string) {
				_ = os.RemoveAll(path)
				_ = os.Chdir(currentDir)
			}(dir, currentDir)

			err = os.Chdir(dir)
			assert.NoError(t, err)
			err = os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte(chart_file), 0644)
			assert.NoError(t, err)
			err = os.WriteFile(filepath.Join(dir, "values.yaml"), []byte(test.values), 0644)
			assert.NoError(t, err)

			app := plural.CreateNewApp(&plural.Plural{Client: nil})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			_, err = captureStdout(app, os.Args)
			assert.NoError(t, err)

			newChart, err := os.ReadFile("./Chart.yaml")
			assert.NoError(t, err)
			assert.Equal(t, test.expectedChart, string(newChart))

		})
	}
}
