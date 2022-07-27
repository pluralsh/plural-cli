package kubernetes

import (
	"context"
	"os"
	"path/filepath"

	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"
	pluralv1alpha1 "github.com/pluralsh/plural-operator/generated/platform/clientset/versioned"
	"github.com/pluralsh/plural/pkg/application"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/pathing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const tokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"

func InKubernetes() bool {
	if os.Getenv("IGNORE_IN_CLUSTER") == "true" {
		return false
	}

	return utils.Exists(tokenFile)
}

type Kube interface {
	Secret(namespace string, name string) (*v1.Secret, error)
	Node(name string) (*v1.Node, error)
	Nodes() (*v1.NodeList, error)
	FinalizeNamespace(namespace string) error
	LogTailList(namespace string) (*v1alpha1.LogTailList, error)
	LogTail(namespace string, name string) (*v1alpha1.LogTail, error)
	ProxyList(namespace string) (*v1alpha1.ProxyList, error)
	Proxy(namespace string, name string) (*v1alpha1.Proxy, error)
	GetClient() *kubernetes.Clientset
}

type kube struct {
	Kube        *kubernetes.Clientset
	Plural      *pluralv1alpha1.Clientset
	Application *application.ApplicationV1Beta1Client
	Dynamic     dynamic.Interface
}

func KubeConfig() (*rest.Config, error) {
	if InKubernetes() {
		return rest.InClusterConfig()
	}

	homedir, _ := os.UserHomeDir()
	conf := pathing.SanitizeFilepath(filepath.Join(homedir, ".kube", "config"))
	return clientcmd.BuildConfigFromFlags("", conf)
}

func Kubernetes() (Kube, error) {
	conf, err := KubeConfig()
	if err != nil {
		return nil, err
	}

	return buildKubeFromConfig(conf)
}

func buildKubeFromConfig(config *rest.Config) (Kube, error) {
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

	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &kube{Kube: clientset, Plural: plural, Application: app, Dynamic: dyn}, nil
}

func (k *kube) Secret(namespace string, name string) (*v1.Secret, error) {
	return k.Kube.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (k *kube) Node(name string) (*v1.Node, error) {
	return k.Kube.CoreV1().Nodes().Get(context.Background(), name, metav1.GetOptions{})
}

func (k *kube) Nodes() (*v1.NodeList, error) {
	return k.Kube.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
}

func (k *kube) FinalizeNamespace(namespace string) error {
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

func (k *kube) LogTailList(namespace string) (*v1alpha1.LogTailList, error) {
	ctx := context.Background()
	return k.Plural.PlatformV1alpha1().LogTails(namespace).List(ctx, metav1.ListOptions{})
}

func (k *kube) LogTail(namespace string, name string) (*v1alpha1.LogTail, error) {
	ctx := context.Background()
	return k.Plural.PlatformV1alpha1().LogTails(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (k *kube) ProxyList(namespace string) (*v1alpha1.ProxyList, error) {
	ctx := context.Background()
	return k.Plural.PlatformV1alpha1().Proxies(namespace).List(ctx, metav1.ListOptions{})
}

func (k *kube) Proxy(namespace string, name string) (*v1alpha1.Proxy, error) {
	ctx := context.Background()
	return k.Plural.PlatformV1alpha1().Proxies(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (k *kube) GetClient() *kubernetes.Clientset {
	return k.Kube
}
