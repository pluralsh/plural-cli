package api

import "github.com/pluralsh/gqlclient"

func (client *client) GetTfProviders() ([]string, error) {
	resp, err := client.pluralClient.GetTfProviders(client.ctx)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0)
	for _, provider := range resp.TerraformProviders {
		result = append(result, string(*provider))
	}

	return result, nil
}

func (client *client) GetTfProviderScaffold(name, version string) (string, error) {
	resp, err := client.pluralClient.GetTfProviderScaffold(client.ctx, gqlclient.Provider(name), &version)
	if err != nil {
		return "", err
	}

	return *resp.TerraformProvider.Content, err
}
