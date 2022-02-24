package api

var TfProvidersQuery = `
	query {
		terraformProviders
	}
`

var TfProviderQuery = `
	query Provider($name: Provider!) {
		terraformProvider(name: $name) {
			name
			content
		}
	}
`

func (client *Client) GetTfProviders() ([]string, error) {
	var resp struct {
		TerraformProviders []string
	}
	req := client.Build(TfProvidersQuery)
	err := client.Run(req, &resp)
	return resp.TerraformProviders, err
}

func (client *Client) GetTfProviderScaffold(name string) (string, error) {
	var resp struct {
		TerraformProvider struct {
			Name    string
			Content string
		}
	}
	req := client.Build(TfProviderQuery)
	req.Var("name", name)
	err := client.Run(req, &resp)
	return resp.TerraformProvider.Content, err
}
