package plural_test

import (
	"os"
	"testing"

	consoleclient "github.com/pluralsh/console-client-go"
	plural "github.com/pluralsh/plural-cli/cmd/plural"
	"github.com/pluralsh/plural-cli/pkg/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
				Client:            nil,
				ConsoleClient:     client,
				Kube:              nil,
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
				Client:            nil,
				ConsoleClient:     client,
				Kube:              nil,
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
				Client:            nil,
				ConsoleClient:     client,
				Kube:              nil,
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
				Client:            nil,
				ConsoleClient:     client,
				Kube:              nil,
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
