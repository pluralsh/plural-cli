package bundle_test

import (
	"os"
	"testing"

	"github.com/pluralsh/plural-cli/pkg/manifest"

	pluralclient "github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"

	"github.com/pluralsh/plural-cli/cmd/command/plural"
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/test/mocks"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v2"
)

func TestBundleList(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		recipe           []*api.Recipe
		expectedResponse string
	}{
		{
			name: `test "bundle list"`,
			args: []string{plural.ApplicationName, "bundle", "list", "test"},
			recipe: []*api.Recipe{
				{
					Id:          "123",
					Name:        "test",
					Provider:    "aws",
					Description: "test application",
				},
			},
			expectedResponse: `+------+------------------+----------+--------------------------------+
| NAME |   DESCRIPTION    | PROVIDER |        INSTALL COMMAND         |
+------+------------------+----------+--------------------------------+
| test | test application | aws      | plural bundle install test     |
|      |                  |          | test                           |
+------+------------------+----------+--------------------------------+
`,
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
			_, err = git.Init()
			assert.NoError(t, err)

			data, err := yaml.Marshal(manifest.ProjectManifest{
				Cluster:  "test",
				Bucket:   "test",
				Project:  "test",
				Provider: "test",
				Region:   "test",
			})
			assert.NoError(t, err)
			err = os.WriteFile("workspace.yaml", data, os.FileMode(0755))
			assert.NoError(t, err)

			client := mocks.NewClient(t)
			client.On("ListRecipes", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(test.recipe, nil)
			app := plural.CreateNewApp(&plural.Plural{Plural: pluralclient.Plural{Client: client}})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			res, err := common.CaptureStdout(app, os.Args)
			assert.NoError(t, err)

			assert.Equal(t, test.expectedResponse, res)
		})
	}
}

func TestBundleInstallNoGitRootDirectory(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedResponse string
	}{
		{
			name:             `test "bundle install" when no root directory`,
			args:             []string{plural.ApplicationName, "bundle", "install", "repo-test", "bundle-test"},
			expectedResponse: `You must run this command at the root of your git repository`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewClient(t)
			app := plural.CreateNewApp(&plural.Plural{Plural: pluralclient.Plural{Client: client}})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			_, err := common.CaptureStdout(app, os.Args)

			assert.Error(t, err)
			assert.Equal(t, test.expectedResponse, err.Error())
		})
	}
}

func TestBundleInstall(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		recipe           *api.Recipe
		expectedResponse string
	}{
		{
			name: `test "bundle install"`,
			args: []string{plural.ApplicationName, "bundle", "install", "repo-test", "bundle-test"},
			recipe: &api.Recipe{
				Id:          "123",
				Name:        "test",
				Provider:    "aws",
				Description: "test application",
				RecipeSections: []*api.RecipeSection{
					{
						Id: "456",
						Repository: &api.Repository{
							Id:          "",
							Name:        "bootstrap",
							Description: "test bootstrap repo",
						},
						RecipeItems:   nil,
						Configuration: nil,
					},
				},
			},
			expectedResponse: "\x1b[2J\x1b[H test bootstrap repo\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create temp environment
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
			_, err = git.Init()
			assert.NoError(t, err)

			client := mocks.NewClient(t)
			client.On("GetRecipe", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(test.recipe, nil)
			client.On("InstallRecipe", mock.AnythingOfType("string")).Return(nil)
			app := plural.CreateNewApp(&plural.Plural{Plural: pluralclient.Plural{Client: client}})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			res, err := common.CaptureStdout(app, os.Args)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedResponse, res)
		})
	}
}
