package cluster

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	clusterapi "sigs.k8s.io/cluster-api/api/v1beta1"
)

type ClusterInterface interface {
	List(ctx context.Context, opts metav1.ListOptions) (*clusterapi.ClusterList, error)
	Get(ctx context.Context, name string, options metav1.GetOptions) (*clusterapi.Cluster, error)
	Create(ctx context.Context, clust *clusterapi.Cluster) (*clusterapi.Cluster, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	// ...
}

type clusterClient struct {
	restClient rest.Interface
	ns         string
}

func (c *clusterClient) List(ctx context.Context, opts metav1.ListOptions) (*clusterapi.ClusterList, error) {
	result := clusterapi.ClusterList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("clusters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *clusterClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*clusterapi.Cluster, error) {
	result := clusterapi.Cluster{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("clusters").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *clusterClient) Create(ctx context.Context, cluster *clusterapi.Cluster) (*clusterapi.Cluster, error) {
	result := clusterapi.Cluster{}
	err := c.restClient.
		Post().
		Namespace(c.ns).
		Resource("clusters").
		Body(cluster).
		Do(ctx).
		Into(&result)

	return &result, err
}

func (c *clusterClient) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.restClient.
		Get().
		Namespace(c.ns).
		Resource("clusters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(ctx)
}

func WatchNamespace(ctx context.Context, client ClusterInterface) (watch.Interface, error) {
	clusters, err := client.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	resourceVersion := clusters.ListMeta.ResourceVersion
	return client.Watch(ctx, metav1.ListOptions{ResourceVersion: resourceVersion})
}
