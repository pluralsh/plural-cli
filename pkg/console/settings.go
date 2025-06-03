package console

import (
	"fmt"

	gqlclient "github.com/pluralsh/console-client-go"
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
