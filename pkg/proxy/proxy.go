package proxy

import (
	"fmt"

	"github.com/michaeljguarino/forge/pkg/types/v1alpha1"
	"github.com/michaeljguarino/forge/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func List(namespace string) (*v1alpha1.ProxyList, error) {
	kube, err := utils.Kubernetes()
	if err != nil {
		return nil, err
	}
	return kube.Forge.Proxies(namespace).List(metav1.ListOptions{})
}

func Exec(namespace string, name string) error {
	kube, err := utils.Kubernetes()
	if err != nil {
		return err
	}
	proxy, err := kube.Forge.Proxies(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	t := proxy.Spec.Type
	switch t {
	case "db":
		secret, err := fetchSecret(namespace, kube, &proxy.Spec.Credentials)
		if err != nil {
			return err
		}
		conn, err := buildConnection(secret, proxy)
		if err != nil {
			return err
		}
		return conn.Connect(namespace)
	case "sh":
		return execShell(namespace, proxy)
	default:
		return fmt.Errorf("Unhandled proxy type %s", t)
	}
}
