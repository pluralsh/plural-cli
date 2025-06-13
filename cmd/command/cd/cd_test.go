package cd_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	consoleclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/cmd/command/plural"
	pluralclient "github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/urfave/cli"
)

func TestListCDClusters(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedResponse string
		result           *consoleclient.ListClusters
	}{
		{
			name:             `test "deployments clusters list" when returns nil object`,
			result:           nil,
			args:             []string{plural.ApplicationName, "deployments", "clusters", "list"},
			expectedResponse: `returned objects list [ListClusters] is nil`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewConsoleClient(t)
			client.On("ListClusters").Return(test.result, nil)
			app := plural.CreateNewApp(&plural.Plural{
				Plural: pluralclient.Plural{
					ConsoleClient: client,
				},
				HelmConfiguration: nil,
			})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			_, err := captureStdout(app, os.Args)

			assert.Error(t, err)
			assert.Equal(t, test.expectedResponse, err.Error())
		})
	}
}

func TestDescribeCDCluster(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedResponse string
		expectedError    string
		result           *consoleclient.ClusterFragment
	}{
		{
			name:          `test "deployments clusters describe" when returns nil`,
			result:        nil,
			args:          []string{plural.ApplicationName, "deployments", "clusters", "describe", "abc"},
			expectedError: `existing cluster is empty`,
		},
		{
			name: `test "deployments clusters describe"`,
			result: &consoleclient.ClusterFragment{
				ID:   "abc",
				Name: "test",
			},
			args: []string{plural.ApplicationName, "deployments", "clusters", "describe", "abc"},
			expectedResponse: `Id:    abc
Name:  test
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewConsoleClient(t)
			client.On("GetCluster", mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).Return(test.result, nil)
			app := plural.CreateNewApp(&plural.Plural{
				Plural: pluralclient.Plural{
					ConsoleClient: client,
				},
				HelmConfiguration: nil,
			})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			out, err := captureStdout(app, os.Args)

			if test.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, test.expectedError, err.Error())
			}
			if test.expectedResponse != "" {
				assert.Equal(t, test.expectedResponse, out)
			}
		})
	}
}

func TestListCDRepositories(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedResponse string
		expectedError    string
		result           *consoleclient.ListGitRepositories
	}{
		{
			name:          `test "deployments repositories list" when returns nil`,
			result:        nil,
			args:          []string{plural.ApplicationName, "deployments", "repositories", "list"},
			expectedError: `returned objects list [ListRepositories] is nil`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewConsoleClient(t)
			client.On("ListRepositories").Return(test.result, nil)
			app := plural.CreateNewApp(&plural.Plural{
				Plural: pluralclient.Plural{
					ConsoleClient: client,
				},
				HelmConfiguration: nil,
			})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			out, err := captureStdout(app, os.Args)

			if test.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, test.expectedError, err.Error())
			}
			if test.expectedResponse != "" {
				assert.Equal(t, test.expectedResponse, out)
			}
		})
	}
}

func TestListCDServices(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedResponse string
		expectedError    string
		result           []*consoleclient.ServiceDeploymentEdgeFragment
	}{
		{
			name:          `test "deployments services list" when returns nil`,
			result:        nil,
			args:          []string{plural.ApplicationName, "deployments", "services", "list", "clusterID"},
			expectedError: `returned objects list [ListClusterServices] is nil`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewConsoleClient(t)
			client.On("ListClusterServices", mock.AnythingOfType("*string"), mock.AnythingOfType("*string")).Return(test.result, nil)
			app := plural.CreateNewApp(&plural.Plural{
				Plural: pluralclient.Plural{
					ConsoleClient: client,
				},
				HelmConfiguration: nil,
			})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			out, err := captureStdout(app, os.Args)

			if test.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, test.expectedError, err.Error())
			}
			if test.expectedResponse != "" {
				assert.Equal(t, test.expectedResponse, out)
			}
		})
	}
}

func TestPipelineTrigger(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError string
		result        *consoleclient.PipelineContextFragment
	}{
		{
			name: "test trigger pipeline with context",
			args: []string{plural.ApplicationName, "cd", "pipelines", "trigger", "pipeline-123", "--context", `{"key":"value"}`},
			result: &consoleclient.PipelineContextFragment{
				ID: "context-123",
			},
		},
		{
			name:          "test trigger pipeline without context",
			args:          []string{plural.ApplicationName, "cd", "pipelines", "trigger", "pipeline-123"},
			expectedError: "Required flag \"context\" not set",
		},
		{
			name:          "test trigger pipeline with empty context",
			args:          []string{plural.ApplicationName, "cd", "pipelines", "trigger", "pipeline-123", "--context", ""},
			expectedError: "no context provided",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewConsoleClient(t)
			if test.result != nil {
				client.On("CreatePipelineContext", mock.AnythingOfType("string"), mock.AnythingOfType("client.PipelineContextAttributes")).Return(test.result, nil)
			}

			app := plural.CreateNewApp(&plural.Plural{
				Plural: pluralclient.Plural{
					ConsoleClient: client,
				},
				HelmConfiguration: nil,
			})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			err := app.Run(os.Args)

			if test.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, test.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func captureStdout(app *cli.App, arg []string) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := app.Run(arg)
	if err != nil {
		return "", err
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return "", err
	}
	return buf.String(), nil
}
