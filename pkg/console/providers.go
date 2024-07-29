package console

import (
	consoleclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/api"
)

func (c *consoleClient) ListProviders() (*consoleclient.ListProviders, error) {

	result, err := c.client.ListProviders(c.ctx)
	if err != nil {
		return nil, api.GetErrorResponse(err, "ListProviders")
	}

	return result, nil
}

func (c *consoleClient) CreateProviderCredentials(name string, attr consoleclient.ProviderCredentialAttributes) (*consoleclient.CreateProviderCredential, error) {

	result, err := c.client.CreateProviderCredential(c.ctx, attr, name)
	if err != nil {
		return nil, api.GetErrorResponse(err, "CreateProviderCredential")
	}

	return result, nil
}

func (c *consoleClient) DeleteProviderCredentials(id string) (*consoleclient.DeleteProviderCredential, error) {

	result, err := c.client.DeleteProviderCredential(c.ctx, id)
	if err != nil {
		return nil, api.GetErrorResponse(err, "DeleteProviderCredential")
	}

	return result, nil
}

func (c *consoleClient) CreateProvider(attr consoleclient.ClusterProviderAttributes) (*consoleclient.CreateClusterProvider, error) {
	result, err := c.client.CreateClusterProvider(c.ctx, attr)
	if err != nil {
		return nil, api.GetErrorResponse(err, "CreateProvider")
	}

	return result, nil
}
