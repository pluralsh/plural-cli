package api

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
