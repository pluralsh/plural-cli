package console

import (
	"fmt"

	consoleclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/samber/lo"
)

func (c *consoleClient) GetProject(name string) (*consoleclient.ProjectFragment, error) {
	result, err := c.client.GetProject(c.ctx, nil, lo.ToPtr(name))
	if err != nil {
		return nil, api.GetErrorResponse(err, "GetProject")
	}

	return result.Project, nil
}

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

func (c *consoleClient) AgentUrl(id string) (string, error) {
	res, err := c.client.GetAgentURL(c.ctx, id)
	if err != nil {
		return "", err
	}

	if res == nil {
		return "", fmt.Errorf("cluster not found")
	}

	return lo.FromPtr(res.Cluster.AgentURL), nil
}

func (c *consoleClient) GetDeployToken(clusterId, clusterName *string) (string, error) {
	res, err := c.client.GetClusterWithToken(c.ctx, clusterId, clusterName)
	if err != nil {
		return "", err
	}

	if res == nil {
		return "", fmt.Errorf("cluster not found")
	}

	return lo.FromPtr(res.Cluster.DeployToken), nil
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

func (c *consoleClient) DetachCluster(id string) error {
	_, err := c.client.DetachCluster(c.ctx, id)
	return api.GetErrorResponse(err, "DetachCluster")
}

func (c *consoleClient) CreateCluster(attributes consoleclient.ClusterAttributes) (*consoleclient.CreateCluster, error) {
	newCluster, err := c.client.CreateCluster(c.ctx, attributes)
	if err != nil {
		return nil, api.GetErrorResponse(err, "CreateCluster")
	}
	return newCluster, nil
}

func (c *consoleClient) MyCluster() (*consoleclient.MyCluster, error) {
	return c.client.MyCluster(c.ctx)
}
