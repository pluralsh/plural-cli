package utils

import (
	"os"
	"path/filepath"
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/rest"
	"github.com/pluralsh/plural/pkg/application"
	pluralv1alpha1 "github.com/pluralsh/plural-operator/generated/platform/clientset/versioned"
)

const tokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"

func InKubernetes() bool {
	return Exists(tokenFile)
}

type Kube struct {
	Kube  *kubernetes.Clientset
	Plural *pluralv1alpha1.Clientset
	Application *application.ApplicationV1Beta1Client
}

func InClusterKubernetes() (*Kube, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return buildKubeFromConfig(config)
}

func Kubernetes() (*Kube, error) {
	if InKubernetes() {
		return InClusterKubernetes()
	}

	homedir, _ := os.UserHomeDir()
	conf := filepath.Join(homedir, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", conf)
	if err != nil {
		return nil, err
	}

	return buildKubeFromConfig(config)
}

func buildKubeFromConfig(config *rest.Config) (*Kube, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	plural, err := pluralv1alpha1.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	app, err := application.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Kube{Kube: clientset, Plural: plural, Application: app}, nil
}

func (k *Kube) Secret(namespace string, name string) (*v1.Secret, error) {
	return k.Kube.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (k *Kube) Node(name string) (*v1.Node, error) {
	return k.Kube.CoreV1().Nodes().Get(context.Background(), name, metav1.GetOptions{})
}

func (k *Kube) FinalizeNamespace(namespace string) error {
	ctx := context.Background()
	client := k.Kube.CoreV1().Namespaces()
	ns, err := client.Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return err
	}

	ns.Spec.Finalizers = []v1.FinalizerName{}
	_, err = client.Finalize(ctx, ns, metav1.UpdateOptions{})
	return err
}
