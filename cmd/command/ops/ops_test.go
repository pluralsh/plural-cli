package ops_test

import (
	"os"
	"testing"

	"github.com/pluralsh/plural-cli/pkg/common"

	"github.com/pluralsh/plural-cli/cmd/command/plural"
	clientcmd "github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/test/mocks"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v2"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestListNodes(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		nodes            *v1.NodeList
		expectedResponse string
	}{
		{
			name: `test "ops cluster"`,
			args: []string{plural.ApplicationName, "ops", "cluster"},
			nodes: &v1.NodeList{
				Items: []v1.Node{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "cluster-1"},
						Spec:       v1.NodeSpec{},
						Status:     v1.NodeStatus{},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "cluster-2"},
						Spec:       v1.NodeSpec{},
						Status:     v1.NodeStatus{},
					},
				},
			},
			expectedResponse: `+-----------+-----+--------+--------+------+
|   NAME    | CPU | MEMORY | REGION | ZONE |
+-----------+-----+--------+--------+------+
| cluster-1 |   0 |      0 |        |      |
| cluster-2 |   0 |      0 |        |      |
+-----------+-----+--------+--------+------+
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewClient(t)
			kube := mocks.NewKube(t)
			kube.On("Nodes").Return(test.nodes, nil)
			app := plural.CreateNewApp(&plural.Plural{
				Plural: clientcmd.Plural{
					Client: client,
					Kube:   kube,
				},
			})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			res, err := common.CaptureStdout(app, os.Args)
			assert.NoError(t, err)

			assert.Equal(t, test.expectedResponse, res)
		})
	}
}

func TestTerminate(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		node             *v1.Node
		pm               manifest.ProjectManifest
		expectedResponse string
	}{
		{
			name: `test "ops terminate"`,
			args: []string{plural.ApplicationName, "ops", "terminate"},
			node: &v1.Node{

				ObjectMeta: metav1.ObjectMeta{Name: "cluster-1"},
				Spec:       v1.NodeSpec{},
				Status:     v1.NodeStatus{},
			},
			pm: manifest.ProjectManifest{
				Cluster:  "test",
				Bucket:   "test",
				Project:  "test",
				Provider: "test",
				Region:   "test",
			},
			expectedResponse: ``,
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
			data, err := yaml.Marshal(test.pm)
			assert.NoError(t, err)
			err = os.WriteFile("workspace.yaml", data, os.FileMode(0755))
			assert.NoError(t, err)

			client := mocks.NewClient(t)
			kube := mocks.NewKube(t)
			kube.On("Node", mock.AnythingOfType("string")).Return(test.node, nil)
			app := plural.CreateNewApp(&plural.Plural{
				Plural: clientcmd.Plural{
					Client: client,
					Kube:   kube,
				},
			})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			res, err := common.CaptureStdout(app, os.Args)
			assert.NoError(t, err)

			assert.Equal(t, test.expectedResponse, res)
		})
	}
}
