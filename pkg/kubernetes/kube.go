package kubernetes

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/pathing"
	platformv1alpha1 "github.com/pluralsh/plural-operator/apis/platform/v1alpha1"
	pluralv1alpha1 "github.com/pluralsh/plural-operator/generated/client/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/scheme"
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
	SecretList(namespace string, opts metav1.ListOptions) (*v1.SecretList, error)
	SecretCreate(namespace string, secret *v1.Secret) (*v1.Secret, error)
	SecretDelete(namespace string, secretName string) error
	SecretDeleteCollection(namespace string, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Node(name string) (*v1.Node, error)
	Nodes() (*v1.NodeList, error)
	FinalizeNamespace(namespace string) error
	LogTailList(namespace string) (*platformv1alpha1.LogTailList, error)
	LogTail(namespace string, name string) (*platformv1alpha1.LogTail, error)
	ProxyList(namespace string) (*platformv1alpha1.ProxyList, error)
	Proxy(namespace string, name string) (*platformv1alpha1.Proxy, error)
	Apply(path string, force bool) error
	CreateNamespace(namespace string, managedByPlural bool) error
	GetClient() *kubernetes.Clientset
	GetRestClient() *rest.RESTClient
}

var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

type kube struct {
	Kube       *kubernetes.Clientset
	Plural     *pluralv1alpha1.Clientset
	Dynamic    dynamic.Interface
	Discovery  discovery.DiscoveryInterface
	Mapper     *restmapper.DeferredDiscoveryRESTMapper
	RestClient *rest.RESTClient
}

func (k *kube) GetRestClient() *rest.RESTClient {
	return k.RestClient
}

func KubeConfig() (*rest.Config, error) {
	return KubeConfigWithContext("")
}

func KubeConfigWithContext(context string) (*rest.Config, error) {
	if os.Getenv("KUBECONFIG") != "" {
		conf := os.Getenv("KUBECONFIG")
		if len(context) > 0 {
			return buildConfigFromFlags(context, conf)
		}

		return clientcmd.BuildConfigFromFlags("", conf)
	}

	if InKubernetes() {
		return rest.InClusterConfig()
	}

	homedir, _ := os.UserHomeDir()
	conf := pathing.SanitizeFilepath(filepath.Join(homedir, ".kube", "config"))

	if len(context) > 0 {
		return buildConfigFromFlags(context, conf)
	}

	return clientcmd.BuildConfigFromFlags("", conf)
}

func Kubernetes() (Kube, error) {
	conf, err := KubeConfig()
	if err != nil {
		return nil, err
	}

	return buildKubeFromConfig(conf)
}

func buildConfigFromFlags(context, kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{CurrentContext: context}).ClientConfig()
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

	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	config.APIPath = "/api"
	config.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}
	config.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}
	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, err
	}

	return &kube{Kube: clientset, Plural: plural, Dynamic: dyn, Discovery: dc, Mapper: mapper, RestClient: restClient}, nil
}

func (k *kube) Secret(namespace string, name string) (*v1.Secret, error) {
	return k.Kube.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

func (k *kube) SecretList(namespace string, opts metav1.ListOptions) (*v1.SecretList, error) {
	return k.Kube.CoreV1().Secrets(namespace).List(context.Background(), opts)
}

func (k *kube) SecretCreate(namespace string, secret *v1.Secret) (*v1.Secret, error) {
	return k.Kube.CoreV1().Secrets(namespace).Create(context.Background(), secret, metav1.CreateOptions{})
}

func (k *kube) SecretDelete(namespace string, secretName string) error {
	return k.Kube.CoreV1().Secrets(namespace).Delete(context.Background(), secretName, metav1.DeleteOptions{})
}

func (k *kube) SecretDeleteCollection(namespace string, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return k.Kube.CoreV1().Secrets(namespace).DeleteCollection(context.Background(), opts, listOpts)
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

func (k *kube) LogTailList(namespace string) (*platformv1alpha1.LogTailList, error) {
	ctx := context.Background()
	return k.Plural.PlatformV1alpha1().LogTails(namespace).List(ctx, metav1.ListOptions{})
}

func (k *kube) LogTail(namespace string, name string) (*platformv1alpha1.LogTail, error) {
	ctx := context.Background()
	return k.Plural.PlatformV1alpha1().LogTails(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (k *kube) ProxyList(namespace string) (*platformv1alpha1.ProxyList, error) {
	ctx := context.Background()
	return k.Plural.PlatformV1alpha1().Proxies(namespace).List(ctx, metav1.ListOptions{})
}

func (k *kube) Proxy(namespace string, name string) (*platformv1alpha1.Proxy, error) {
	ctx := context.Background()
	return k.Plural.PlatformV1alpha1().Proxies(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (k *kube) GetClient() *kubernetes.Clientset {
	return k.Kube
}

func (k *kube) Apply(path string, force bool) error {
	ctx := context.Background()
	yamlFile, err := utils.ReadFile(path)
	if err != nil {
		return err
	}
	multidocReader := utilyaml.NewYAMLReader(bufio.NewReader(bytes.NewReader([]byte(yamlFile))))
	for {
		buf, err := multidocReader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		// yaml starts with `---`
		if strings.TrimSpace(string(buf)) == "" {
			continue
		}
		obj := &unstructured.Unstructured{}
		_, gvk, err := decUnstructured.Decode(buf, nil, obj)
		if err != nil {
			return err
		}
		mapping, err := k.Mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return err
		}
		var dr dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			// namespaced resources should specify the namespace
			dr = k.Dynamic.Resource(mapping.Resource).Namespace(obj.GetNamespace())
		} else {
			// for cluster-wide resources
			dr = k.Dynamic.Resource(mapping.Resource)
		}

		if _, err := dr.Apply(ctx, obj.GetName(), obj, metav1.ApplyOptions{Force: force, FieldManager: "application/apply-patch"}); err != nil {
			return err
		}
	}

	return nil
}

func (k *kube) CreateNamespace(namespace string, managedByPlural bool) error {
	ctx := context.Background()
	_, err := k.Kube.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		return nil
	}
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"app.plural.sh/name": namespace,
			},
		},
	}
	if managedByPlural {
		ns.Labels["app.kubernetes.io/managed-by"] = "plural"
	}
	_, err = k.Kube.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})

	return err
}
