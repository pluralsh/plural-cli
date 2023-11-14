package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/provider"
	pluraltest "github.com/pluralsh/plural-cli/pkg/test"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
	"github.com/pluralsh/plural-cli/pkg/wkspace"
)

func TestBuildChartValues(t *testing.T) {
	tests := []struct {
		name                  string
		workspace             *wkspace.Workspace
		existingValues        string
		expectedDefaultValues string
		expectedValues        string
		man                   *manifest.ProjectManifest
		expectError           bool
	}{
		{
			name: `test build values first time`,
			expectedDefaultValues: `plrl:
  license: abc
test:
  extraEnv:
  - name: ARM_USE_MSI
    value: "true"
`,
			expectedValues: "",
			man: &manifest.ProjectManifest{
				Cluster:  "test",
				Bucket:   "test",
				Project:  "test",
				Provider: "kind",
				Region:   "test",
			},
			workspace: &wkspace.Workspace{
				Installation: &api.Installation{
					Id: "123",
					Repository: &api.Repository{
						Name: "test",
					},
					LicenseKey: "abc",
				},
				Charts: []*api.ChartInstallation{
					{
						Id: "123",
						Chart: &api.Chart{
							Id:   "123",
							Name: "test",
						},
						Version: &api.Version{
							ValuesTemplate: `output = {
		extraEnv={
			{
				name="ARM_USE_MSI",
				value = 'true'
	
			},
    	}
}`,
							TemplateType: "LUA",
						},
					},
				},
				Context: &manifest.Context{
					Configuration: map[string]map[string]interface{}{
						"test": {},
					},
				},
			},
		},
		{
			name: `test build values when values.yaml exists, add and override variables`,
			expectedDefaultValues: `plrl:
  license: abc
test:
  extraEnv:
  - name: ARM_USE_MSI
    value: "true"
`,
			expectedValues: `plrl:
  license: abc
test:
  enabled: false
  extraEnv:
  - name: TEST
    value: "false"
`,
			existingValues: `plrl:
  license: abc
test:
  enabled: false
  extraEnv:
  - name: TEST
    value: "false"
`,
			man: &manifest.ProjectManifest{
				Cluster:  "test",
				Bucket:   "test",
				Project:  "test",
				Provider: "kind",
				Region:   "test",
			},
			workspace: &wkspace.Workspace{
				Installation: &api.Installation{
					Id: "123",
					Repository: &api.Repository{
						Name: "test",
					},
					LicenseKey: "abc",
				},
				Charts: []*api.ChartInstallation{
					{
						Id: "123",
						Chart: &api.Chart{
							Id:   "123",
							Name: "test",
						},
						Version: &api.Version{
							ValuesTemplate: `output = {
		extraEnv={
			{
				name="ARM_USE_MSI",
				value = 'true'
	
			},
    	}
}`,
							TemplateType: "LUA",
						},
					},
				},
				Context: &manifest.Context{
					Configuration: map[string]map[string]interface{}{
						"test": {},
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir, err := os.MkdirTemp("", "config")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			os.Setenv("HOME", dir)
			defer os.Unsetenv("HOME")

			err = os.Chdir(dir)
			assert.NoError(t, err)

			data, err := yaml.Marshal(test.man)
			assert.NoError(t, err)
			err = os.WriteFile("workspace.yaml", data, os.FileMode(0755))
			assert.NoError(t, err)

			err = os.WriteFile("values.yaml", []byte(test.existingValues), os.FileMode(0755))
			assert.NoError(t, err)

			defaultConfig := pluraltest.GenDefaultConfig()
			err = defaultConfig.Save(config.ConfigName)
			assert.NoError(t, err)

			_, err = git.Init()
			assert.NoError(t, err)
			_, err = git.GitRaw("config", "--global", "user.email", "test@plural.com")
			assert.NoError(t, err)
			_, err = git.GitRaw("config", "--global", "user.name", "test")
			assert.NoError(t, err)
			_, err = git.GitRaw("add", "-A")
			assert.NoError(t, err)
			_, err = git.GitRaw("commit", "-m", "init")
			assert.NoError(t, err)
			_, err = git.GitRaw("remote", "add", "origin", "git@git.test.com:portfolio/space.space_name.git")
			assert.NoError(t, err)

			provider, err := provider.FromManifest(test.man)
			if err != nil {
				t.Fatal(err)
			}
			test.workspace.Provider = provider
			scaffold := Scaffold{
				Name: "test",
				Root: dir,
			}

			err = scaffold.buildChartValues(test.workspace)
			if test.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			defaultValues, err := os.ReadFile(filepath.Join(dir, "default-values.yaml"))
			assert.NoError(t, err)
			assert.Equal(t, test.expectedDefaultValues, string(defaultValues))

			values, err := os.ReadFile(filepath.Join(dir, "values.yaml"))
			assert.NoError(t, err)
			assert.Equal(t, test.expectedValues, string(values))
		})
	}
}
