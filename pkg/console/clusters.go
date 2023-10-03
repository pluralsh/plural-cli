package console

import (
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

func (c *consoleClient) GetCluster(id string) (*consoleclient.GetCluster, error) {

	result, err := c.client.GetCluster(c.ctx, &id)
	if err != nil {
		return nil, api.GetErrorResponse(err, "GetCluster")
	}

	return result, nil
}

func (c *consoleClient) UpdateCluster(id string, attr consoleclient.ClusterUpdateAttributes) (*consoleclient.UpdateCluster, error) {

	result, err := c.client.UpdateCluster(c.ctx, id, attr)
	if err != nil {
		return nil, api.GetErrorResponse(err, "UpdateCluster")
	}

	return result, nil
}
