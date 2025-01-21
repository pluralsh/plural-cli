package console

import (
	consoleclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/api"
)

func (c *consoleClient) CreateClusterRegistration(attributes consoleclient.ClusterRegistrationCreateAttributes) (*consoleclient.ClusterRegistrationFragment, error) {
	response, err := c.client.CreateClusterRegistration(c.ctx, attributes)
	if err != nil {
		return nil, api.GetErrorResponse(err, "CreateClusterRegistration")
	}
	return response.CreateClusterRegistration, nil
}

func (c *consoleClient) IsClusterRegistrationComplete(machineID string) (bool, *consoleclient.ClusterRegistrationFragment) {
	response, err := c.client.GetClusterRegistration(c.ctx, nil, &machineID)
	if err != nil {
		return false, nil // TODO: Log error?
	}

	return response.ClusterRegistration.Name != nil, response.ClusterRegistration
}
