package machinepool

import (
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	clusterapiExp "sigs.k8s.io/cluster-api/exp/api/v1beta1"
)

type MachinePoolV1Beta1Interface interface {
	MachinePools(namespace string) MachinePoolInterface
}

type MachinePoolV1Beta1Client struct {
	restClient rest.Interface
}

func NewForConfig(c *rest.Config) (*MachinePoolV1Beta1Client, error) {
	if err := AddToScheme(scheme.Scheme); err != nil {
		return nil, err
	}

	config := *c
	config.ContentConfig.GroupVersion = &clusterapiExp.GroupVersion
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &MachinePoolV1Beta1Client{restClient: client}, nil
}

func (c *MachinePoolV1Beta1Client) MachinePools(namespace string) MachinePoolInterface {
	return &machinepoolClient{
		restClient: c.restClient,
		ns:         namespace,
	}
}
