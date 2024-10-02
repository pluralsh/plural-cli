package api_test

import (
	"os"
	"testing"

	"github.com/pluralsh/plural-cli/cmd/command/plural"
	"github.com/pluralsh/plural-cli/pkg/api"
	pluralclient "github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListArtifacts(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		artifacts        []api.Artifact
		expectedResponse string
		expectedError    string
	}{
		{
			name: `test "api list artifacts" with single response`,
			args: []string{plural.ApplicationName, "api", "list", "artifacts", "test"},
			artifacts: []api.Artifact{{
				Id:       "abc",
				Name:     "test",
				Blob:     "test",
				Sha:      "xyz",
				Platform: "aws",
			}},
			expectedResponse: `+-----+------+----------+------+-----+
| ID  | NAME | PLATFORM | BLOB | SHA |
+-----+------+----------+------+-----+
| abc | test | aws      | test | xyz |
+-----+------+----------+------+-----+
`,
		},
		{
			name:          `test "api list artifacts" without {repository-id} parameter`,
			args:          []string{plural.ApplicationName, "api", "list", "artifacts"},
			expectedError: "Not enough arguments provided: needs {repository-id}. Try running --help to see usage.",
			artifacts:     []api.Artifact{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewClient(t)
			if test.expectedError == "" {
				client.On("ListArtifacts", mock.AnythingOfType("string")).Return(test.artifacts, nil)
			}
			app := plural.CreateNewApp(&plural.Plural{Plural: pluralclient.Plural{
				Client: client,
			}})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			res, err := common.CaptureStdout(app, os.Args)
			if test.expectedError != "" {
				assert.Equal(t, err.Error(), test.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedResponse, res)
			}

		})
	}
}

func TestGetInstallations(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		installations    []*api.Installation
		expectedResponse string
	}{
		{
			name: `test "api list installations"`,
			args: []string{plural.ApplicationName, "api", "list", "installations"},
			installations: []*api.Installation{
				{Id: "123", Repository: &api.Repository{Id: "abc", Name: "test-1", Publisher: &api.Publisher{Name: "Plural"}}},
				{Id: "456", Repository: &api.Repository{Id: "def", Name: "test-2", Publisher: &api.Publisher{Name: "Plural"}}},
			},
			expectedResponse: `+------------+---------------+-----------+
| REPOSITORY | REPOSITORY ID | PUBLISHER |
+------------+---------------+-----------+
| test-1     | abc           | Plural    |
| test-2     | def           | Plural    |
+------------+---------------+-----------+
`,
		},
		{
			name:          `test "api list installations" when Repository is nil`,
			args:          []string{plural.ApplicationName, "api", "list", "installations"},
			installations: []*api.Installation{{Id: "abc"}},
			expectedResponse: `+------------+---------------+-----------+
| REPOSITORY | REPOSITORY ID | PUBLISHER |
+------------+---------------+-----------+
+------------+---------------+-----------+
`,
		},
		{
			name:          `test "api list installations" when Publisher is nil`,
			args:          []string{plural.ApplicationName, "api", "list", "installations"},
			installations: []*api.Installation{{Id: "abc", Repository: &api.Repository{Id: "abc", Name: "test"}}},
			expectedResponse: `+------------+---------------+-----------+
| REPOSITORY | REPOSITORY ID | PUBLISHER |
+------------+---------------+-----------+
| test       | abc           |           |
+------------+---------------+-----------+
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewClient(t)
			client.On("GetInstallations").Return(test.installations, nil)
			app := plural.CreateNewApp(&plural.Plural{Plural: pluralclient.Plural{
				Client: client,
			}})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			res, err := common.CaptureStdout(app, os.Args)
			assert.NoError(t, err)

			assert.Equal(t, test.expectedResponse, res)
		})
	}
}

func TestGetCharts(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		charts           []*api.Chart
		expectedResponse string
		expectedError    string
	}{
		{
			name: `test "api list charts" with single response`,
			args: []string{plural.ApplicationName, "api", "list", "charts", "test"},
			charts: []*api.Chart{{
				Id:            "123",
				Name:          "test",
				Description:   "test chart",
				LatestVersion: "0.1.0",
			}},
			expectedResponse: `+-----+------+-------------+----------------+
| ID  | NAME | DESCRIPTION | LATEST VERSION |
+-----+------+-------------+----------------+
| 123 | test | test chart  | 0.1.0          |
+-----+------+-------------+----------------+
`,
		},
		{
			name:          `test "api list charts" without {repository-id} parameter`,
			args:          []string{plural.ApplicationName, "api", "list", "charts"},
			charts:        []*api.Chart{},
			expectedError: "Not enough arguments provided: needs {repository-id}. Try running --help to see usage.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewClient(t)
			if test.expectedError == "" {
				client.On("GetCharts", mock.AnythingOfType("string")).Return(test.charts, nil)
			}
			app := plural.CreateNewApp(&plural.Plural{Plural: pluralclient.Plural{
				Client: client,
			}})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			res, err := common.CaptureStdout(app, os.Args)
			if test.expectedError != "" {
				assert.Equal(t, err.Error(), test.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedResponse, res)
			}
		})
	}
}

func TestGetTerraform(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		terraform        []*api.Terraform
		expectedResponse string
		expectedError    string
	}{
		{
			name: `test "api list terraform"`,
			args: []string{plural.ApplicationName, "api", "list", "terraform", "test"},
			terraform: []*api.Terraform{
				{
					Id:          "123",
					Name:        "test-1",
					Description: "test terraform",
				},
				{
					Id:          "456",
					Name:        "test-2",
					Description: "test terraform",
				},
			},
			expectedResponse: `+-----+--------+----------------+
| ID  |  NAME  |  DESCRIPTION   |
+-----+--------+----------------+
| 123 | test-1 | test terraform |
| 456 | test-2 | test terraform |
+-----+--------+----------------+
`,
		},
		{
			name:          `test "api list terraform" without {repository-id} parameter`,
			args:          []string{plural.ApplicationName, "api", "list", "terraform"},
			terraform:     []*api.Terraform{},
			expectedError: "Not enough arguments provided: needs {repository-id}. Try running --help to see usage.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewClient(t)
			if test.expectedError == "" {
				client.On("GetTerraform", mock.AnythingOfType("string")).Return(test.terraform, nil)
			}

			app := plural.CreateNewApp(&plural.Plural{Plural: pluralclient.Plural{
				Client: client,
			}})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			res, err := common.CaptureStdout(app, os.Args)
			if test.expectedError != "" {
				assert.Equal(t, err.Error(), test.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedResponse, res)
			}
		})
	}
}

func TestGetVersons(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		versions         []*api.Version
		expectedResponse string
		expectedError    string
	}{
		{
			name: `test "api list versions"`,
			args: []string{plural.ApplicationName, "api", "list", "versions", "abc"},
			versions: []*api.Version{
				{
					Id:      "abc",
					Version: "1",
				},
				{
					Id:      "abc",
					Version: "2",
				},
			},
			expectedResponse: `+-----+---------+
| ID  | VERSION |
+-----+---------+
| abc |       1 |
| abc |       2 |
+-----+---------+
`,
		},
		{
			name:          `test "api list versions" without {chart-id} parameter`,
			args:          []string{plural.ApplicationName, "api", "list", "versions"},
			versions:      []*api.Version{},
			expectedError: "Not enough arguments provided: needs {chart-id}. Try running --help to see usage.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewClient(t)
			if test.expectedError == "" {
				client.On("GetVersions", mock.AnythingOfType("string")).Return(test.versions, nil)
			}
			app := plural.CreateNewApp(&plural.Plural{Plural: pluralclient.Plural{
				Client: client,
			}})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			res, err := common.CaptureStdout(app, os.Args)
			if test.expectedError != "" {
				assert.Equal(t, err.Error(), test.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedResponse, res)
			}
		})
	}
}

func TestGetChartInstallations(t *testing.T) {
	tests := []struct {
		name               string
		args               []string
		chartInstallations []*api.ChartInstallation
		expectedResponse   string
		expectedError      string
	}{
		{
			name: `test "api list chartinstallations"`,
			args: []string{plural.ApplicationName, "api", "list", "chartinstallations", "abc"},
			chartInstallations: []*api.ChartInstallation{
				{
					Id: "abc",
					Chart: &api.Chart{
						Id:   "abc",
						Name: "test-1",
					},
					Version: &api.Version{
						Version: "1",
					},
				},
				{
					Id: "abc",
					Chart: &api.Chart{
						Id:   "abc",
						Name: "test-2",
					},
					Version: &api.Version{
						Version: "2",
					},
				},
			},
			expectedResponse: `+-----+----------+------------+---------+
| ID  | CHART ID | CHART NAME | VERSION |
+-----+----------+------------+---------+
| abc | abc      | test-1     |       1 |
| abc | abc      | test-2     |       2 |
+-----+----------+------------+---------+
`,
		},
		{
			name:               `test "api list chartinstallations" without {repository-id} parameter`,
			args:               []string{plural.ApplicationName, "api", "list", "chartinstallations"},
			chartInstallations: []*api.ChartInstallation{},
			expectedError:      "Not enough arguments provided: needs {repository-id}. Try running --help to see usage.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewClient(t)
			if test.expectedError == "" {
				client.On("GetChartInstallations", mock.AnythingOfType("string")).Return(test.chartInstallations, nil)
			}
			app := plural.CreateNewApp(&plural.Plural{Plural: pluralclient.Plural{
				Client: client,
			}})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			res, err := common.CaptureStdout(app, os.Args)
			if test.expectedError != "" {
				assert.Equal(t, err.Error(), test.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedResponse, res)
			}
		})
	}
}

func TestGetTerraformInstallations(t *testing.T) {
	tests := []struct {
		name                   string
		args                   []string
		terraformInstallations []*api.TerraformInstallation
		expectedResponse       string
		expectedError          string
	}{
		{
			name: `test "api list terraforminstallations"`,
			args: []string{plural.ApplicationName, "api", "list", "terraforminstallations", "abc"},
			terraformInstallations: []*api.TerraformInstallation{
				{
					Id: "abc",
					Terraform: &api.Terraform{
						Id:   "cde",
						Name: "tf-1",
					},
				},
				{
					Id: "abc",
					Terraform: &api.Terraform{
						Id:   "fgh",
						Name: "tf-2",
					},
				},
			},
			expectedResponse: `+-----+--------------+------+
| ID  | TERRAFORM ID | NAME |
+-----+--------------+------+
| abc | cde          | tf-1 |
| abc | fgh          | tf-2 |
+-----+--------------+------+
`,
		},
		{
			name:                   `test "api list terraforminstallations" without {repository-id} parameter`,
			args:                   []string{plural.ApplicationName, "api", "list", "terraforminstallations"},
			terraformInstallations: []*api.TerraformInstallation{},
			expectedError:          "Not enough arguments provided: needs {repository-id}. Try running --help to see usage.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewClient(t)
			if test.expectedError == "" {
				client.On("GetTerraformInstallations", mock.AnythingOfType("string")).Return(test.terraformInstallations, nil)
			}
			app := plural.CreateNewApp(&plural.Plural{Plural: pluralclient.Plural{
				Client: client,
			}})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			res, err := common.CaptureStdout(app, os.Args)
			if test.expectedError != "" {
				assert.Equal(t, err.Error(), test.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedResponse, res)
			}
		})
	}
}
