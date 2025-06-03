package vpn

import (
	"github.com/pluralsh/plural-cli/pkg/kubernetes"
	"github.com/pluralsh/plural-operator/apis/vpn/v1alpha1"

	v1 "k8s.io/api/core/v1"
)

func ListServers(kube kubernetes.Kube, namespace string) (*v1alpha1.WireguardServerList, error) {
	return kube.WireguardServerList(namespace)
}

func GetServer(kube kubernetes.Kube, namespace string, name string) (*v1alpha1.WireguardServer, error) {
	return kube.WireguardServer(namespace, name)
}

func ListPeers(kube kubernetes.Kube, namespace string) (*v1alpha1.WireguardPeerList, error) {
	return kube.WireguardPeerList(namespace)
}

func GetPeer(kube kubernetes.Kube, namespace string, name string) (*v1alpha1.WireguardPeer, error) {
	return kube.WireguardPeer(namespace, name)
}

func GetPeerConfigSecret(kube kubernetes.Kube, namespace string, name string) (*v1.Secret, error) {
	return kube.Secret(namespace, name)
}

func CreatePeer(kube kubernetes.Kube, namespace string, peer *v1alpha1.WireguardPeer) (*v1alpha1.WireguardPeer, error) {
	return kube.WireguardPeerCreate(namespace, peer)
}

func DeletePeer(kube kubernetes.Kube, namespace string, name string) error {
	return kube.WireguardPeerDelete(namespace, name)
}
