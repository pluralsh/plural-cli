package application

import (
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/application/api/v1beta1"

	// "k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

type ApplicationV1Beta1Interface interface {
	Applications(namespace string) ApplicationInterface
}

type ApplicationV1Beta1Client struct {
	restClient rest.Interface
}

func NewForConfig(c *rest.Config) (*ApplicationV1Beta1Client, error) {
	if err := AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}

	config := *c
	config.ContentConfig.GroupVersion = &v1beta1.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &ApplicationV1Beta1Client{restClient: client}, nil
}

func (c *ApplicationV1Beta1Client) Applications(namespace string) ApplicationInterface {
	return &applicationClient{
		restClient: c.restClient,
		ns:         namespace,
	}
}
