package main_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
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
			// create temp environment
			dir, err := ioutil.TempDir("", "config")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

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

func TestValidate(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		installations    []*api.Installation
		charts           []*api.ChartInstallation
		tfs              []*api.TerraformInstallation
		pm               manifest.ProjectManifest
		ctx              manifest.Context
		expectedResponse string
	}{
		{
			name: `test "validate"`,
			args: []string{plural.ApplicationName, "validate"},
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
			pm: manifest.ProjectManifest{
				Cluster:  "test",
				Bucket:   "test",
				Project:  "test",
				Provider: "kind",
				Region:   "test",
			},
			ctx: manifest.Context{
				Bundles: []*manifest.Bundle{
					{
						Repository: "cde",
						Name:       "cde",
					},
					{
						Repository: "abc",
						Name:       "abc",
					},
				},
			},
			tfs:    []*api.TerraformInstallation{},
			charts: []*api.ChartInstallation{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create temp environment
			dir, err := ioutil.TempDir("", "config")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			err = os.Chdir(dir)
			assert.NoError(t, err)

			data, err := yaml.Marshal(test.pm)
			assert.NoError(t, err)
			err = os.WriteFile("workspace.yaml", data, os.FileMode(0755))
			assert.NoError(t, err)

			data, err = yaml.Marshal(test.ctx)
			assert.NoError(t, err)
			err = os.WriteFile("context.yaml", data, os.FileMode(0755))
			assert.NoError(t, err)

			client := mocks.NewClient(t)
			client.On("GetInstallations").Return(test.installations, nil)
			client.On("GetPackageInstallations", mock.AnythingOfType("string")).Return(test.charts, test.tfs, nil)
			app := plural.CreateNewApp(&plural.Plural{Client: client})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			resp, err := captureStdout(app, os.Args)
			assert.NoError(t, err)
			assert.Equal(t, resp, test.expectedResponse)
		})
	}
}
