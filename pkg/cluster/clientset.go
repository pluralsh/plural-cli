package cluster

import (
	"k8s.io/client-go/kubernetes/scheme"
	clusterapi "sigs.k8s.io/cluster-api/api/v1beta1"

	// "k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
)

type ClusterV1Beta1Interface interface {
	Clusters(namespace string) ClusterInterface
}

type ClusterV1Beta1Client struct {
	restClient rest.Interface
}

func NewForConfig(c *rest.Config) (*ClusterV1Beta1Client, error) {
	if err := AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}

	config := *c
	config.ContentConfig.GroupVersion = &clusterapi.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &ClusterV1Beta1Client{restClient: client}, nil
}

func (c *ClusterV1Beta1Client) Clusters(namespace string) ClusterInterface {
	return &clusterClient{
		restClient: c.restClient,
		ns:         namespace,
	}
}
