package workspace_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/pluralsh/plural-cli/pkg/common"

	"github.com/pluralsh/plural-cli/cmd/command/plural"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
)

const subchart = "subchart"

func TestHelmCommands(t *testing.T) {
	// create temp environment
	currentDir, err := os.Getwd()
	assert.NoError(t, err)
	dir, err := os.MkdirTemp("", "config")
	assert.NoError(t, err)
	defer func(path, currentDir string) {
		_ = os.RemoveAll(path)
		_ = os.Chdir(currentDir)
	}(dir, currentDir)
	tFiles, err := filepath.Abs("../../../pkg/test/helm")
	assert.NoError(t, err)
	err = utils.CopyDir(tFiles, filepath.Join(dir, subchart))
	assert.NoError(t, err)
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

	tests := []struct {
		name           string
		args           []string
		expectedOutput string
		store          *storage.Storage
		directory      string
	}{
		{
			name:           `test helm-template`,
			args:           []string{plural.ApplicationName, "workspace", "helm-template", subchart},
			expectedOutput: "subchart/helm/output/template.txt",
			store:          storageFixture(),
			directory:      dir,
		},
		{
			name:      `test helm install`,
			args:      []string{plural.ApplicationName, "workspace", "helm", subchart},
			store:     storageFixture(),
			directory: filepath.Join(dir, subchart),
		},
		{
			name:      `test helm upgrade`,
			args:      []string{plural.ApplicationName, "workspace", "helm", subchart},
			store:     storageReleaseDeployed(t),
			directory: filepath.Join(dir, subchart),
		},
		{
			name:           `test helm-diff`,
			args:           []string{plural.ApplicationName, "workspace", "helm-diff", subchart},
			expectedOutput: "subchart/helm/output/diff.txt",
			store:          storageReleaseDeployed(t),
			directory:      dir,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actionConfig := &action.Configuration{
				Releases:     test.store,
				KubeClient:   &kubefake.PrintingKubeClient{Out: io.Discard},
				Capabilities: chartutil.DefaultCapabilities,
				Log:          func(format string, v ...interface{}) {},
			}
			err = os.Chdir(test.directory)
			assert.NoError(t, err)
			defer func() {
				err := os.Chdir(dir)
				assert.NoError(t, err)
			}()

			app := plural.CreateNewApp(&plural.Plural{HelmConfiguration: actionConfig})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			output, err := common.CaptureStdout(app, os.Args)
			assert.NoError(t, err)
			if test.expectedOutput != "" {
				expected, err := utils.ReadFile(test.expectedOutput)
				assert.NoError(t, err)
				assert.Equal(t, expected, output)
			}
		})
	}
}

func storageFixture() *storage.Storage {
	return storage.Init(driver.NewMemory())
}

func storageReleaseDeployed(t *testing.T) *storage.Storage {
	fixture := storageFixture()
	err := fixture.Create(&release.Release{
		Name: "subchart",
		Info: &release.Info{Status: release.StatusDeployed},
		Chart: &chart.Chart{
			Metadata: &chart.Metadata{
				Name:    "Myrelease-Chart",
				Version: "1.2.3",
			},
		},
		Version: 1,
	})
	if err != nil {
		t.Fatal("can't create storage")
	}
	return fixture
}
