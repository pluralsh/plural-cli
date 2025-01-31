package console

import (
	"errors"

	consoleclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/api"
)

func (c *consoleClient) CreateBootstrapToken(attributes consoleclient.BootstrapTokenAttributes) (string, error) {
	response, err := c.client.CreateBootstrapToken(c.ctx, attributes)
	if err != nil {
		return "", api.GetErrorResponse(err, "CreateBootstrapToken")
	}
	if response.CreateBootstrapToken == nil {
		return "", errors.New("CreateBootstrapToken returned nil response")
	}

	return response.CreateBootstrapToken.Token, nil
}

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
		return false, nil
	}

	return response.ClusterRegistration.Name != nil, response.ClusterRegistration
}
