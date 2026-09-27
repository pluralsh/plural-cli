package console

import (
	"fmt"

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

func (c *consoleClient) GetProvider(id string) (*consoleclient.ClusterProviderFragment, error) {
	response, err := c.client.GetClusterProvider(c.ctx, id)
	if err != nil {
		return nil, api.GetErrorResponse(err, "GetClusterProvider")
	}
	if response == nil || response.ClusterProvider == nil {
		return nil, fmt.Errorf("cluster provider %s was not found", id)
	}
	return response.ClusterProvider, nil
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
