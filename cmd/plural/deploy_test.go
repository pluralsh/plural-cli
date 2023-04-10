package plural_test

import (
	"os"
	"testing"

	"gopkg.in/yaml.v2"

	plural "github.com/pluralsh/plural/cmd/plural"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/test/mocks"
	"github.com/stretchr/testify/assert"
)

func TestBuildContext(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		installations    []*api.Installation
		expectedResponse string
	}{
		{
			name: `test "build-context"`,
			args: []string{plural.ApplicationName, "build-context"},
			installations: []*api.Installation{{
				Id: "abc",
				Repository: &api.Repository{
					Id:   "abc",
					Name: "abc",
				},
			},
				{
					Id: "cde",
					Repository: &api.Repository{
						Id:   "cde",
						Name: "cde",
					},
				},
			},
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

			client := mocks.NewClient(t)
			client.On("GetInstallations").Return(test.installations, nil)
			app := plural.CreateNewApp(&plural.Plural{Client: client})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			_, err = captureStdout(app, os.Args)
			assert.NoError(t, err)

			dat, err := os.ReadFile(dir + "/context.yaml")
			assert.NoError(t, err)
			ctx := manifest.VersionedContext{}
			err = yaml.Unmarshal(dat, &ctx)
			assert.NoError(t, err)
			for _, inst := range test.installations {
				_, ok := ctx.Spec.Configuration[inst.Repository.Name]
				assert.Equal(t, true, ok, "expected configuration for repository name %s", inst.Repository.Name)
			}

		})
	}
}
