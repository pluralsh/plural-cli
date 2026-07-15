package console

import (
	"fmt"
	"strings"

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
	settings, err := c.getGlobalSettings()
	if err != nil && IsUnknownGraphQLField(err, "agentHelmValuesTemplateable") {
		return c.GetGlobalSettingsMinimal()
	}
	return settings, err
}

func (c *consoleClient) getGlobalSettings() (*gqlclient.DeploymentSettingsFragment, error) {
	resp, err := c.client.GetDeploymentSettings(c.ctx)
	if err != nil {
		return nil, api.GetErrorResponse(err, "GetDeploymentSettings")
	}
	if resp == nil {
		return nil, fmt.Errorf("returned GetDeploymentSettings object is nil")
	}
	return resp.DeploymentSettings, nil
}

func (c *consoleClient) GetGlobalSettingsMinimal() (*gqlclient.DeploymentSettingsFragment, error) {
	resp, err := c.client.GetDeploymentSettingsMinimal(c.ctx)
	if err != nil {
		return nil, api.GetErrorResponse(err, "GetDeploymentSettingsMinimal")
	}
	if resp == nil {
		return nil, fmt.Errorf("returned GetDeploymentSettingsMinimal object is nil")
	}
	return toDeploymentSettingsFragment(resp.DeploymentSettings), nil
}

func toDeploymentSettingsFragment(minimal *gqlclient.DeploymentSettingsMinimalFragment) *gqlclient.DeploymentSettingsFragment {
	if minimal == nil {
		return nil
	}

	return &gqlclient.DeploymentSettingsFragment{
		AgentHelmValues: minimal.AgentHelmValues,
		AgentVsn:        minimal.AgentVsn,
	}
}

func IsUnknownGraphQLField(err error, field string) bool {
	if err == nil {
		return false
	}

	msg := err.Error()
	quotedField := `"` + field + `"`
	return strings.Contains(msg, "Cannot query field") && strings.Contains(msg, quotedField)
}
