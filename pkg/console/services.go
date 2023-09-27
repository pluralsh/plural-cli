package console

import (
	gqlclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/plural/pkg/api"
)

func (c *consoleClient) ListClusterServices(clusterId string) (*gqlclient.ListServiceDeployment, error) {
	result, err := c.client.ListServiceDeployment(c.ctx, nil, nil, nil, &clusterId)
	if err != nil {
		return nil, api.GetErrorResponse(err, "ListClusterServices")
	}

	return result, nil
}

func (c *consoleClient) CreateClusterService(clusterId string, attributes gqlclient.ServiceDeploymentAttributes) (*gqlclient.CreateServiceDeployment, error) {
	result, err := c.client.CreateServiceDeployment(c.ctx, clusterId, attributes)
	if err != nil {
		return nil, api.GetErrorResponse(err, "CreateServiceDeployment")
	}

	return result, nil
}

func (c *consoleClient) UpdateClusterService(serviceId string, attributes gqlclient.ServiceUpdateAttributes) (*gqlclient.UpdateServiceDeployment, error) {
	result, err := c.client.UpdateServiceDeployment(c.ctx, serviceId, attributes)
	if err != nil {
		return nil, api.GetErrorResponse(err, "UpdateClusterService")
	}

	return result, nil
}

func (c *consoleClient) GetClusterService(serviceId string) (*gqlclient.GetServiceDeployment, error) {
	result, err := c.client.GetServiceDeployment(c.ctx, serviceId)
	if err != nil {
		return nil, api.GetErrorResponse(err, "UpdateClusterService")
	}

	return result, nil
}
