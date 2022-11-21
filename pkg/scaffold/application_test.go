package scaffold_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pluralsh/plural/pkg/scaffold"
	"github.com/stretchr/testify/assert"
)

const values_file = `console:
  enabled: true
  ingress:
    annotations:
      external-dns.alpha.kubernetes.io/target: 127.0.0.1
    console_dns: console.onplural.sh
  license: abc
  provider: kind
`

func TestListRepositories(t *testing.T) {
	tests := []struct {
		name             string
		appName          string
		expectedResponse map[string]interface{}
	}{
		{
			name:    `test HelmValues`,
			appName: "test",
			expectedResponse: map[string]interface{}{
				"console": map[string]interface{}{
					"enabled": true,
					"ingress": map[string]interface{}{
						"annotations": map[string]interface{}{
							"external-dns.alpha.kubernetes.io/target": "127.0.0.1",
						},
						"console_dns": "console.onplural.sh",
					},
					"license":  "abc",
					"provider": "kind",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir, err := os.MkdirTemp("", "config")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			dirPath := filepath.Join(dir, test.appName, "helm", test.appName)
			err = os.MkdirAll(dirPath, os.ModePerm)
			assert.NoError(t, err)
			err = os.WriteFile(filepath.Join(dirPath, "values.yaml"), []byte(values_file), 0644)
			assert.NoError(t, err)

			application := scaffold.Applications{
				Root: dir,
			}
			res, err := application.HelmValues("test")

			assert.NoError(t, err)
			assert.Equal(t, test.expectedResponse, res)
		})
	}
}
