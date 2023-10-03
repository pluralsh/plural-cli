package console

import (
	"fmt"
	gqlclient "github.com/pluralsh/console-client-go"

	"github.com/pluralsh/plural/pkg/api"
)

func (c *consoleClient) ListClusterServices(clusterId, cluster *string) ([]*gqlclient.ServiceDeploymentEdgeFragment, error) {
	if clusterId == nil && cluster == nil {
		return nil, fmt.Errorf("clusterId and cluster can not be null")
	}
	if clusterId != nil {
		result, err := c.client.ListServiceDeployment(c.ctx, nil, nil, nil, clusterId)
		if err != nil {
			return nil, api.GetErrorResponse(err, "ListServiceDeployment")
		}
		if result == nil {
			return nil, fmt.Errorf("the result from ListServiceDeployment is null")
		}
		return result.ServiceDeployments.Edges, nil
	}
	result, err := c.client.ListServiceDeploymentByHandle(c.ctx, nil, nil, nil, cluster)
	if err != nil {
		return nil, api.GetErrorResponse(err, "ListServiceDeploymentByHandle")
	}
	if result == nil {
		return nil, fmt.Errorf("the result from ListServiceDeploymentByHandle is null")
	}
	return result.ServiceDeployments.Edges, nil
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

func (c *consoleClient) GetClusterService(serviceId, serviceName, clusterName *string) (*gqlclient.ServiceDeploymentExtended, error) {
	if serviceId == nil && serviceName == nil && clusterName == nil {
		return nil, fmt.Errorf("serviceId, serviceName and clusterName can not be null")
	}
	if serviceId != nil {
		result, err := c.client.GetServiceDeployment(c.ctx, *serviceId)
		if err != nil {
			return nil, api.GetErrorResponse(err, "GetClusterService")
		}
		if result == nil {
			return nil, fmt.Errorf("the result from GetServiceDeployment is null")
		}
		return result.ServiceDeployment, nil
	}
	result, err := c.client.GetServiceDeploymentByHandle(c.ctx, *clusterName, *serviceName)
	if err != nil {
		return nil, api.GetErrorResponse(err, "GetServiceDeploymentByHandle")
	}
	if result == nil {
		return nil, fmt.Errorf("the result from GetServiceDeploymentByHandle is null")
	}

	return result.ServiceDeployment, nil
}

func (c *consoleClient) DeleteClusterService(serviceId string) (*gqlclient.DeleteServiceDeployment, error) {
	result, err := c.client.DeleteServiceDeployment(c.ctx, serviceId)
	if err != nil {
		return nil, api.GetErrorResponse(err, "DeleteClusterService")
	}

	return result, nil
}
