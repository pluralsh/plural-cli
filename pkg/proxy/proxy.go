package proxy

import (
	"fmt"
	"context"

	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"
	"github.com/pluralsh/plural/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func List(namespace string) (*v1alpha1.ProxyList, error) {
	kube, err := utils.Kubernetes()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return kube.Plural.PlatformV1alpha1().Proxies(namespace).List(ctx, metav1.ListOptions{})
}

func Exec(namespace string, name string) error {
	kube, err := utils.Kubernetes()
	if err != nil {
		return err
	}

	ctx := context.Background()
	proxy, err := kube.Plural.PlatformV1alpha1().Proxies(namespace).Get(ctx, name, metav1.GetOptions{})
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
