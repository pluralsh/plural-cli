package application

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/application/api/v1beta1"
)

type ApplicationInterface interface {
	List(ctx context.Context, opts metav1.ListOptions) (*v1beta1.ApplicationList, error)
	Get(ctx context.Context, name string, options metav1.GetOptions) (*v1beta1.Application, error)
	Create(ctx context.Context, app *v1beta1.Application) (*v1beta1.Application, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	// ...
}

type applicationClient struct {
	restClient rest.Interface
	ns         string
}

func (c *applicationClient) List(ctx context.Context, opts metav1.ListOptions) (*v1beta1.ApplicationList, error) {
	result := v1beta1.ApplicationList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("applications").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *applicationClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1beta1.Application, error) {
	result := v1beta1.Application{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("applications").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *applicationClient) Create(ctx context.Context, app *v1beta1.Application) (*v1beta1.Application, error) {
	result := v1beta1.Application{}
	err := c.restClient.
		Post().
		Namespace(c.ns).
		Resource("applications").
		Body(app).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *applicationClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.ns).
		Resource("applications").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(ctx)
}

func WatchNamespace(ctx context.Context, client ApplicationInterface) (watch.Interface, error) {
	apps, err := client.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	resourceVersion := apps.ListMeta.ResourceVersion
	return client.Watch(ctx, metav1.ListOptions{ResourceVersion: resourceVersion})
}
