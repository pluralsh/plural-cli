package machinepool

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	clusterapiExp "sigs.k8s.io/cluster-api/exp/api/v1beta1"
)

type MachinePoolInterface interface {
	List(ctx context.Context, opts metav1.ListOptions) (*clusterapiExp.MachinePoolList, error)
	Get(ctx context.Context, name string, options metav1.GetOptions) (*clusterapiExp.MachinePool, error)
	Create(ctx context.Context, mp *clusterapiExp.MachinePool) (*clusterapiExp.MachinePool, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Update(ctx context.Context, mp *clusterapiExp.MachinePool) (*clusterapiExp.MachinePool, error)
	// ...
}

type machinepoolClient struct {
	restClient rest.Interface
	ns         string
}

func (c *machinepoolClient) List(ctx context.Context, opts metav1.ListOptions) (*clusterapiExp.MachinePoolList, error) {
	result := clusterapiExp.MachinePoolList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("machinepools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *machinepoolClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*clusterapiExp.MachinePool, error) {
	result := clusterapiExp.MachinePool{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("machinepools").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *machinepoolClient) Update(ctx context.Context, mp *clusterapiExp.MachinePool) (*clusterapiExp.MachinePool, error) {
	result := clusterapiExp.MachinePool{}
	err := c.restClient.
		Put().
		Namespace(c.ns).
		Resource("machinepools").
		Name(mp.Name).
		Body(mp).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *machinepoolClient) Create(ctx context.Context, machinepool *clusterapiExp.MachinePool) (*clusterapiExp.MachinePool, error) {
	result := clusterapiExp.MachinePool{}
	err := c.restClient.
		Post().
		Namespace(c.ns).
		Resource("machinepools").
		Body(machinepool).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *machinepoolClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.ns).
		Resource("machinepools").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(ctx)
}

func WatchNamespace(ctx context.Context, client MachinePoolInterface, listOps metav1.ListOptions) (watch.Interface, error) {
	mps, err := client.List(ctx, listOps)
	if err != nil {
		return nil, err
	}
	resourceVersion := mps.ListMeta.ResourceVersion
	return client.Watch(ctx, metav1.ListOptions{ResourceVersion: resourceVersion})
}
