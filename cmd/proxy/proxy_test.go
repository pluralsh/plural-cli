package proxy_test

import (
	clientcmd "github.com/pluralsh/plural-cli/pkg/client"
	"os"
	"testing"

	"github.com/pluralsh/plural-cli/pkg/common"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pluralsh/plural-operator/apis/platform/v1alpha1"
	"github.com/stretchr/testify/mock"

	"github.com/pluralsh/plural-cli/cmd/plural"
	"github.com/pluralsh/plural-cli/pkg/test/mocks"
	"github.com/stretchr/testify/assert"
)

func TestProxyList(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		proxyList        *v1alpha1.ProxyList
		expectedResponse string
	}{
		{
			name: `test "proxy list"`,
			args: []string{plural.ApplicationName, "proxy", "list", "test"},
			proxyList: &v1alpha1.ProxyList{
				TypeMeta: metav1.TypeMeta{},
				Items: []v1alpha1.Proxy{
					{
						ObjectMeta: metav1.ObjectMeta{Name: "proxy-1"},
						Spec:       v1alpha1.ProxySpec{Type: v1alpha1.Sh, Target: "test-1"},
					},
					{
						ObjectMeta: metav1.ObjectMeta{Name: "proxy-1"},
						Spec:       v1alpha1.ProxySpec{Type: v1alpha1.Web, Target: "test-2"},
					},
				},
			},
			expectedResponse: `+---------+------+--------+
|  NAME   | TYPE | TARGET |
+---------+------+--------+
| proxy-1 | sh   | test-1 |
| proxy-1 | web  | test-2 |
+---------+------+--------+
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewClient(t)
			kube := mocks.NewKube(t)
			kube.On("ProxyList", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(test.proxyList, nil)
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
