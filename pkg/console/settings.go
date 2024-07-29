package console

import (
	"fmt"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/api"
)

func (c *consoleClient) UpdateDeploymentSettings(attr gqlclient.DeploymentSettingsAttributes) (*gqlclient.UpdateDeploymentSettings, error) {
	resp, err := c.client.UpdateDeploymentSettings(c.ctx, attr)
	if err != nil {
		return nil, api.GetErrorResponse(err, "UpdateDeploymentSettings")
	}
	if resp == nil {
		return nil, fmt.Errorf("returned UpdateDeploymentSettings are nil")
	}

	return resp, nil
}

func (c *consoleClient) GetGlobalSettings() (*gqlclient.DeploymentSettingsFragment, error) {
	resp, err := c.client.GetDeploymentSettings(c.ctx)
	if err != nil {
		return nil, api.GetErrorResponse(err, "GetDeploymentSettings")
	}
	if resp == nil {
		return nil, fmt.Errorf("returned GetDeploymentSettings object is nil")
	}
	return resp.DeploymentSettings, nil
}
