package plural_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/mock"

	plural "github.com/pluralsh/plural/cmd/plural"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/test/mocks"
	"github.com/stretchr/testify/assert"
)

func TestTopSort(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		installations    []*api.Installation
		charts           []*api.ChartInstallation
		tfs              []*api.TerraformInstallation
		expectedResponse string
	}{
		{
			name: `test "topsort"`,
			args: []string{plural.ApplicationName, "topsort"},
			installations: []*api.Installation{
				{
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
			tfs:              []*api.TerraformInstallation{},
			charts:           []*api.ChartInstallation{},
			expectedResponse: "cde\nabc\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewClient(t)
			client.On("GetInstallations").Return(test.installations, nil)
			client.On("GetPackageInstallations", mock.AnythingOfType("string")).Return(test.charts, test.tfs, nil)
			app := plural.CreateNewApp(&plural.Plural{Client: client})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			res, err := captureStdout(app, os.Args)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedResponse, res)
		})
	}
}
