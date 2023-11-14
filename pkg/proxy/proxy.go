package proxy

import (
	"fmt"

	"github.com/pluralsh/plural-cli/pkg/kubernetes"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-operator/apis/platform/v1alpha1"
)

func List(kube kubernetes.Kube, namespace string) (*v1alpha1.ProxyList, error) {
	return kube.ProxyList(namespace)
}

func Print(l *v1alpha1.ProxyList) error {
	headers := []string{"Name", "Type", "Target"}
	return utils.PrintTable[v1alpha1.Proxy](l.Items, headers, func(p v1alpha1.Proxy) ([]string, error) {
		return []string{p.Name, string(p.Spec.Type), p.Spec.Target}, nil
	})
}

func Exec(kube kubernetes.Kube, namespace string, name string) error {
	proxy, err := kube.Proxy(namespace, name)
	if err != nil {
		return err
	}

	t := proxy.Spec.Type
	switch t {
	case v1alpha1.Db:
		secret, err := fetchSecret(namespace, kube, proxy.Spec.Credentials)
		if err != nil {
			return err
		}
		conn, err := buildConnection(secret, proxy)
		if err != nil {
			return err
		}
		return conn.Connect(namespace)
	case v1alpha1.Sh:
		return execShell(namespace, proxy)
	case v1alpha1.Web:
		return execWeb(namespace, proxy, kube)
	default:
		return fmt.Errorf("Unhandled proxy type %s", t)
	}
}
