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
	return true, &consoleclient.ClusterRegistrationFragment{} // TODO: Check if ClusterRegistration already has name assigned.
}
