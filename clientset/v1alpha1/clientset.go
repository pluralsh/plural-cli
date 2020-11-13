package v1alpha1

import (
	"github.com/michaeljguarino/forge/types/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

type Clientset interface {
	Proxies(namespace string) ProxyInterface
}

type ForgeV1Alpha1Client struct {
	restClient rest.Interface
	scheme *runtime.Scheme
}

func NewForConfig(c *rest.Config) (*ForgeV1Alpha1Client, error) {
	sc := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(sc); err != nil {
		return nil, err
	}

	config := *c
	config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: v1alpha1.GroupName, Version: v1alpha1.GroupVersion}
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.NewCodecFactory(sc)
	config.UserAgent = rest.DefaultKubernetesUserAgent()
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &ForgeV1Alpha1Client{restClient: client, scheme: sc}, nil
}

func (c *ForgeV1Alpha1Client) Proxies(namespace string) ProxyInterface {
	return &proxyClient{
		restClient: c.restClient,
		scheme:     c.scheme,
		ns:         namespace,
	}
}
