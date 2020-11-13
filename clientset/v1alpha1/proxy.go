package v1alpha1

import (
	"context"

	"github.com/michaeljguarino/forge/types/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
)

type ProxyInterface interface {
	List(opts metav1.ListOptions) (*v1alpha1.ProxyList, error)
	Get(name string, options metav1.GetOptions) (*v1alpha1.Proxy, error)
	Create(*v1alpha1.Proxy) (*v1alpha1.Proxy, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
}

type proxyClient struct {
	restClient rest.Interface
	ns         string
	scheme     *runtime.Scheme
}

func (c *proxyClient) List(opts metav1.ListOptions) (*v1alpha1.ProxyList, error) {
	result := v1alpha1.ProxyList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("proxies").
		VersionedParams(&opts, runtime.NewParameterCodec(c.scheme)).
		Do(context.Background()).
		Into(&result)
	return &result, err
}

func (c *proxyClient) Get(name string, opts metav1.GetOptions) (*v1alpha1.Proxy, error) {
	result := v1alpha1.Proxy{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("proxies").
		Name(name).
		VersionedParams(&opts, runtime.NewParameterCodec(c.scheme)).
		Do(context.Background()).
		Into(&result)
	return &result, err
}

func (c *proxyClient) Create(prxy *v1alpha1.Proxy) (*v1alpha1.Proxy, error) {
	result := v1alpha1.Proxy{}
	err := c.restClient.
		Post().
		Namespace(c.ns).
		Resource("proxies").
		Body(prxy).
		Do(context.Background()).
		Into(&result)
	return &result, err
}

func (c *proxyClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.ns).
		Resource("proxies").
		VersionedParams(&opts, runtime.NewParameterCodec(c.scheme)).
		Watch(context.Background())
}
