package console

import (
	"fmt"

	consoleclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/plural/pkg/api"
)

func (c *consoleClient) ListClusters() (*consoleclient.ListClusters, error) {

	result, err := c.client.ListClusters(c.ctx, nil, nil, nil)
	if err != nil {
		return nil, api.GetErrorResponse(err, "ListClusters")
	}

	return result, nil
}

func (c *consoleClient) GetCluster(clusterId, clusterName *string) (*consoleclient.ClusterFragment, error) {
	if clusterId == nil && clusterName == nil {
		return nil, fmt.Errorf("clusterId and clusterName can not be null")
	}
	if clusterId != nil {
		result, err := c.client.GetCluster(c.ctx, clusterId)
		if err != nil {
			return nil, api.GetErrorResponse(err, "GetCluster")
		}
		return result.Cluster, nil
	}
	result, err := c.client.GetClusterByHandle(c.ctx, clusterName)
	if err != nil {
		return nil, api.GetErrorResponse(err, "GetCluster")
	}

	return result.Cluster, nil
}

func (c *consoleClient) UpdateCluster(id string, attr consoleclient.ClusterUpdateAttributes) (*consoleclient.UpdateCluster, error) {

	result, err := c.client.UpdateCluster(c.ctx, id, attr)
	if err != nil {
		return nil, api.GetErrorResponse(err, "UpdateCluster")
	}

	return result, nil
}

func (c *consoleClient) DeleteCluster(id string) error {
	_, err := c.client.DeleteCluster(c.ctx, id)
	return api.GetErrorResponse(err, "DeleteCluster")
}
