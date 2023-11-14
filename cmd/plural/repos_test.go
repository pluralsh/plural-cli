package plural_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/pluralsh/plural-cli/cmd/plural"
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/test/mocks"
)

func TestListRepositories(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		repos            []*api.Repository
		expectedResponse string
	}{
		{
			name: `test "repos list"`,
			args: []string{plural.ApplicationName, "repos", "list"},
			repos: []*api.Repository{
				{
					Id:          "123",
					Name:        "test",
					Description: "test application",
					Publisher: &api.Publisher{
						Id:   "456",
						Name: "test",
					},
					Recipes: []*api.Recipe{
						{
							Id:   "789",
							Name: "r1",
						},
						{
							Id:   "101",
							Name: "r2",
						},
					},
				},
			},
			expectedResponse: `+------+------------------+-----------+---------+
| REPO |   DESCRIPTION    | PUBLISHER | BUNDLES |
+------+------------------+-----------+---------+
| test | test application | test      | r1, r2  |
+------+------------------+-----------+---------+
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewClient(t)
			client.On("ListRepositories", mock.AnythingOfType("string")).Return(test.repos, nil)
			app := plural.CreateNewApp(&plural.Plural{Client: client})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			res, err := captureStdout(app, os.Args)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedResponse, res)
		})
	}
}

// func TestResetRepositories(t *testing.T) {
// 	tests := []struct {
// 		name             string
// 		args             []string
// 		count            int
// 		expectedResponse string
// 	}{
// 		{
// 			name:  `test "repos reset"`,
// 			args:  []string{plural.ApplicationName, "repos", "reset"},
// 			count: 5,
// 			expectedResponse: `Deleted 5 installations in app.plural.sh
// (you can recreate these at any time and any running infrastructure is not affected, plural will simply no longer deliver upgrades)
// `,
// 		},
// 	}
// 	for _, test := range tests {
// 		t.Run(test.name, func(t *testing.T) {
// 			client := mocks.NewClient(t)
// 			client.On("ResetInstallations").Return(test.count, nil)
// 			app := plural.CreateNewApp(&plural.Plural{Client: client})
// 			app.HelpName = plural.ApplicationName
// 			os.Args = test.args
// 			res, err := captureStdout(app, os.Args)
// 			assert.NoError(t, err)
// 			assert.Equal(t, test.expectedResponse, res)
// 		})
// 	}
// }
