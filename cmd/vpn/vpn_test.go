package vpn_test

import (
	"os"
	"testing"

	"github.com/pluralsh/plural-cli/cmd/plural"
	"github.com/pluralsh/plural-cli/pkg/api"
	pluralclient "github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/test/mocks"
	vpnv1alpha1 "github.com/pluralsh/plural-operator/apis/vpn/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestServerList(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		servers          *vpnv1alpha1.WireguardServerList
		expectedResponse string
		installation     *api.Installation
	}{
		{
			name: `test "vpn list servers"`,
			args: []string{plural.ApplicationName, "vpn", "list", "servers"},
			servers: &vpnv1alpha1.WireguardServerList{
				Items: []vpnv1alpha1.WireguardServer{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "wireguard",
							Namespace: "wireguard",
						},
						Spec: vpnv1alpha1.WireguardServerSpec{
							WireguardImage: "dkr.plural.sh/bootstrap/wireguard-server:0.1.2",
						},
						Status: vpnv1alpha1.WireguardServerStatus{
							Port:     "51820",
							Hostname: "k8s-wireguar-wireguar-xxxxxxxxxx-xxxxxxxxxxxxxxxx.elb.us-east-1.amazonaws.com",
							Ready:    true,
						},
					},
				},
			},
			installation: &api.Installation{
				Id: "123", Repository: &api.Repository{Id: "abc", Name: "wireguard", Publisher: &api.Publisher{Name: "Plural"}},
			},
			expectedResponse: `+-----------+-------------------------------------------------------------------------------+-------+-------+
|   NAME    |                                   HOSTNAME                                    | PORT  | READY |
+-----------+-------------------------------------------------------------------------------+-------+-------+
| wireguard | k8s-wireguar-wireguar-xxxxxxxxxx-xxxxxxxxxxxxxxxx.elb.us-east-1.amazonaws.com | 51820 | true  |
+-----------+-------------------------------------------------------------------------------+-------+-------+
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewClient(t)
			kube := mocks.NewKube(t)
			client.On("GetInstallation", "wireguard").Return(test.installation, nil)
			kube.On("WireguardServerList", "wireguard").Return(test.servers, nil)
			app := plural.CreateNewApp(&plural.Plural{
				Plural: pluralclient.Plural{
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

func TestClientList(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		peers            *vpnv1alpha1.WireguardPeerList
		expectedResponse string
		installation     *api.Installation
	}{
		{
			name: `test "vpn list clients" without server flag`,
			args: []string{plural.ApplicationName, "vpn", "list", "clients"},
			peers: &vpnv1alpha1.WireguardPeerList{
				Items: []vpnv1alpha1.WireguardPeer{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-client-1",
							Namespace: "wireguard",
						},
						Spec: vpnv1alpha1.WireguardPeerSpec{
							WireguardRef: "wireguard",
							Address:      "10.8.0.2",
							PublicKey:    "test-public-key",
						},
						Status: vpnv1alpha1.WireguardPeerStatus{
							ConfigRef: corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "test-client-1-config",
								},
								Key: "wg0.conf",
							},
							Ready: true,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-client-2",
							Namespace: "wireguard",
						},
						Spec: vpnv1alpha1.WireguardPeerSpec{
							WireguardRef: "wireguard2",
							Address:      "10.8.0.3",
							PublicKey:    "test-public-key",
						},
						Status: vpnv1alpha1.WireguardPeerStatus{
							ConfigRef: corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "test-client-2-config",
								},
								Key: "wg0.conf",
							},
							Ready: false,
						},
					},
				},
			},
			installation: &api.Installation{
				Id: "123", Repository: &api.Repository{Id: "abc", Name: "wireguard", Publisher: &api.Publisher{Name: "Plural"}},
			},
			expectedResponse: `+---------------+----------+----------------------+-----------------+-------+
|     NAME      | ADDRESS  |    CONFIG SECRET     |   PUBLIC KEY    | READY |
+---------------+----------+----------------------+-----------------+-------+
| test-client-1 | 10.8.0.2 | test-client-1-config | test-public-key | true  |
+---------------+----------+----------------------+-----------------+-------+
`,
		},
		{
			name: `test "vpn list clients" with server flag`,
			args: []string{plural.ApplicationName, "vpn", "list", "clients", "--server", "wireguard2"},
			peers: &vpnv1alpha1.WireguardPeerList{
				Items: []vpnv1alpha1.WireguardPeer{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-client-1",
							Namespace: "wireguard",
						},
						Spec: vpnv1alpha1.WireguardPeerSpec{
							WireguardRef: "wireguard",
							Address:      "10.8.0.2",
							PublicKey:    "test-public-key",
						},
						Status: vpnv1alpha1.WireguardPeerStatus{
							ConfigRef: corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "test-client-1-config",
								},
								Key: "wg0.conf",
							},
							Ready: true,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-client-2",
							Namespace: "wireguard",
						},
						Spec: vpnv1alpha1.WireguardPeerSpec{
							WireguardRef: "wireguard2",
							Address:      "10.8.0.3",
							PublicKey:    "test-public-key",
						},
						Status: vpnv1alpha1.WireguardPeerStatus{
							ConfigRef: corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "test-client-2-config",
								},
								Key: "wg0.conf",
							},
							Ready: false,
						},
					},
				},
			},
			installation: &api.Installation{
				Id: "123", Repository: &api.Repository{Id: "abc", Name: "wireguard", Publisher: &api.Publisher{Name: "Plural"}},
			},
			expectedResponse: `+---------------+----------+----------------------+-----------------+-------+
|     NAME      | ADDRESS  |    CONFIG SECRET     |   PUBLIC KEY    | READY |
+---------------+----------+----------------------+-----------------+-------+
| test-client-2 | 10.8.0.3 | test-client-2-config | test-public-key | false |
+---------------+----------+----------------------+-----------------+-------+
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewClient(t)
			kube := mocks.NewKube(t)
			client.On("GetInstallation", "wireguard").Return(test.installation, nil)
			kube.On("WireguardPeerList", "wireguard").Return(test.peers, nil)
			app := plural.CreateNewApp(&plural.Plural{
				Plural: pluralclient.Plural{
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

func TestClientCreate(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		peer             *vpnv1alpha1.WireguardPeer
		expectedResponse string
		installation     *api.Installation
		server           *vpnv1alpha1.WireguardServer
		expectedError    string
	}{
		{
			name: `test "vpn create client" without specifying server`,
			args: []string{plural.ApplicationName, "vpn", "create", "client", "test-client"},
			peer: &vpnv1alpha1.WireguardPeer{

				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-client",
					Namespace: "wireguard",
				},
				Spec: vpnv1alpha1.WireguardPeerSpec{
					WireguardRef: "wireguard",
					Address:      "10.8.0.2",
					PublicKey:    "test-public-key",
				},
				Status: vpnv1alpha1.WireguardPeerStatus{
					ConfigRef: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "test-client-config",
						},
						Key: "wg0.conf",
					},
					Ready: true,
				},
			},
			installation: &api.Installation{
				Id: "123", Repository: &api.Repository{Id: "abc", Name: "wireguard", Publisher: &api.Publisher{Name: "Plural"}},
			},
			server: &vpnv1alpha1.WireguardServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wireguard",
					Namespace: "wireguard",
				},
				Spec: vpnv1alpha1.WireguardServerSpec{
					WireguardImage: "dkr.plural.sh/bootstrap/wireguard-server:0.1.2",
				},
				Status: vpnv1alpha1.WireguardServerStatus{
					Port:     "51820",
					Hostname: "k8s-wireguar-wireguar-xxxxxxxxxx-xxxxxxxxxxxxxxxx.elb.us-east-1.amazonaws.com",
					Ready:    true,
				},
			},
			expectedResponse: `+-------------+----------+-----------+--------------------+-----------------+-------+
|    NAME     | ADDRESS  |  SERVER   |   CONFIG SECRET    |   PUBLIC KEY    | READY |
+-------------+----------+-----------+--------------------+-----------------+-------+
| test-client | 10.8.0.2 | wireguard | test-client-config | test-public-key | true  |
+-------------+----------+-----------+--------------------+-----------------+-------+
`,
		},
		{
			name: `test "vpn create client" with specifying server flag`,
			args: []string{plural.ApplicationName, "vpn", "create", "client", "test-client", "--server", "wireguard2"},
			peer: &vpnv1alpha1.WireguardPeer{

				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-client",
					Namespace: "wireguard",
				},
				Spec: vpnv1alpha1.WireguardPeerSpec{
					WireguardRef: "wireguard2",
					Address:      "10.8.0.2",
					PublicKey:    "test-public-key",
				},
				Status: vpnv1alpha1.WireguardPeerStatus{
					ConfigRef: corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "test-client-config",
						},
						Key: "wg0.conf",
					},
					Ready: true,
				},
			},
			installation: &api.Installation{
				Id: "123", Repository: &api.Repository{Id: "abc", Name: "wireguard", Publisher: &api.Publisher{Name: "Plural"}},
			},
			server: &vpnv1alpha1.WireguardServer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wireguard2",
					Namespace: "wireguard",
				},
				Spec: vpnv1alpha1.WireguardServerSpec{
					WireguardImage: "dkr.plural.sh/bootstrap/wireguard-server:0.1.2",
				},
				Status: vpnv1alpha1.WireguardServerStatus{
					Port:     "51820",
					Hostname: "k8s-wireguar-wireguar-xxxxxxxxxx-xxxxxxxxxxxxxxxx.elb.us-east-1.amazonaws.com",
					Ready:    true,
				},
			},
			expectedResponse: `+-------------+----------+------------+--------------------+-----------------+-------+
|    NAME     | ADDRESS  |   SERVER   |   CONFIG SECRET    |   PUBLIC KEY    | READY |
+-------------+----------+------------+--------------------+-----------------+-------+
| test-client | 10.8.0.2 | wireguard2 | test-client-config | test-public-key | true  |
+-------------+----------+------------+--------------------+-----------------+-------+
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewClient(t)
			kube := mocks.NewKube(t)
			client.On("GetInstallation", "wireguard").Return(test.installation, nil)
			if test.expectedError == "" {
				kube.On("WireguardServer", "wireguard", mock.AnythingOfType("string")).Return(test.server, nil)
				kube.On("WireguardPeerCreate", "wireguard", mock.AnythingOfType("*v1alpha1.WireguardPeer")).Return(test.peer, nil)
			}
			app := plural.CreateNewApp(&plural.Plural{
				Plural: pluralclient.Plural{
					Client: client,
					Kube:   kube,
				},
			})
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
