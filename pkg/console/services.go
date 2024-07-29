package console

import (
	"fmt"

	gqlclient "github.com/pluralsh/console/go/client"

	"github.com/pluralsh/plural-cli/pkg/api"
)

func (c *consoleClient) ListClusterServices(clusterId, clusterName *string) ([]*gqlclient.ServiceDeploymentEdgeFragment, error) {
	if clusterId == nil && clusterName == nil {
		return nil, fmt.Errorf("clusterId and clusterName can not be null")
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
	result, err := c.client.ListServiceDeploymentByHandle(c.ctx, nil, nil, nil, clusterName)
	if err != nil {
		return nil, api.GetErrorResponse(err, "ListServiceDeploymentByHandle")
	}
	if result == nil {
		return nil, fmt.Errorf("the result from ListServiceDeploymentByHandle is null")
	}
	return result.ServiceDeployments.Edges, nil
}

func (c *consoleClient) CreateClusterService(clusterId, clusterName *string, attributes gqlclient.ServiceDeploymentAttributes) (*gqlclient.ServiceDeploymentExtended, error) {
	if clusterId == nil && clusterName == nil {
		return nil, fmt.Errorf("clusterId and clusterName can not be null")
	}
	if clusterId != nil {
		result, err := c.client.CreateServiceDeployment(c.ctx, *clusterId, attributes)
		if err != nil {
			return nil, api.GetErrorResponse(err, "CreateServiceDeployment")
		}
		if result == nil {
			return nil, fmt.Errorf("the result from CreateServiceDeployment is null")
		}
		return result.CreateServiceDeployment, nil
	}

	result, err := c.client.CreateServiceDeploymentWithHandle(c.ctx, *clusterName, attributes)
	if err != nil {
		return nil, api.GetErrorResponse(err, "CreateServiceDeploymentWithHandle")
	}
	if result == nil {
		return nil, fmt.Errorf("the result from CreateServiceDeploymentWithHandle is null")
	}
	return result.CreateServiceDeployment, nil
}

func (c *consoleClient) UpdateClusterService(serviceId, serviceName, clusterName *string, attributes gqlclient.ServiceUpdateAttributes) (*gqlclient.ServiceDeploymentExtended, error) {
	if serviceId == nil && serviceName == nil && clusterName == nil {
		return nil, fmt.Errorf("serviceId, serviceName and clusterName can not be null")
	}
	if serviceId != nil {
		result, err := c.client.UpdateServiceDeployment(c.ctx, *serviceId, attributes)
		if err != nil {
			return nil, api.GetErrorResponse(err, "UpdateClusterService")
		}
		if result == nil {
			return nil, fmt.Errorf("the result from UpdateServiceDeployment is null")
		}

		return result.UpdateServiceDeployment, nil
	}
	result, err := c.client.UpdateServiceDeploymentWithHandle(c.ctx, *clusterName, *serviceName, attributes)
	if err != nil {
		return nil, api.GetErrorResponse(err, "UpdateServiceDeploymentWithHandle")
	}
	if result == nil {
		return nil, fmt.Errorf("the result from UpdateServiceDeploymentWithHandle is null")
	}

	return result.UpdateServiceDeployment, nil
}

func (c *consoleClient) CloneService(clusterId string, serviceId, serviceName, clusterName *string, attributes gqlclient.ServiceCloneAttributes) (*gqlclient.ServiceDeploymentFragment, error) {
	if serviceId == nil && serviceName == nil && clusterName == nil {
		return nil, fmt.Errorf("serviceId, serviceName and clusterName can not be null")
	}
	if serviceId != nil {
		result, err := c.client.CloneServiceDeployment(c.ctx, clusterId, *serviceId, attributes)
		if err != nil {
			return nil, api.GetErrorResponse(err, "CloneService")
		}

		return result.CloneService, nil
	}
	result, err := c.client.CloneServiceDeploymentWithHandle(c.ctx, clusterId, *clusterName, *serviceName, attributes)
	if err != nil {
		return nil, api.GetErrorResponse(err, "CloneServiceWithHandle")
	}

	return result.CloneService, nil
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

func (c *consoleClient) KickClusterService(serviceId, serviceName, clusterName *string) (*gqlclient.ServiceDeploymentExtended, error) {
	if serviceId == nil && serviceName == nil && clusterName == nil {
		return nil, fmt.Errorf("serviceId, serviceName and clusterName can not be null")
	}
	if serviceId != nil {
		result, err := c.client.KickService(c.ctx, *serviceId)
		if err != nil {
			return nil, api.GetErrorResponse(err, "KickService")
		}
		if result == nil {
			return nil, fmt.Errorf("the result from KickService is null")
		}
		return result.KickService, nil
	}
	result, err := c.client.KickServiceByHandle(c.ctx, *clusterName, *serviceName)
	if err != nil {
		return nil, api.GetErrorResponse(err, "KickServiceByHandle")
	}
	if result == nil {
		return nil, fmt.Errorf("the result from KickServiceByHandle is null")
	}

	return result.KickService, nil
}
