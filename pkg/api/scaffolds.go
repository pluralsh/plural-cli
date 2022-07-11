package api

var TfProvidersQuery = `
	query {
		terraformProviders
	}
`

var TfProviderQuery = `
	query Provider($name: Provider!, $vsn: String) {
		terraformProvider(name: $name, vsn: $vsn) {
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

func (client *Client) GetTfProviderScaffold(name, version string) (string, error) {
	var resp struct {
		TerraformProvider struct {
			Name    string
			Content string
		}
	}
	req := client.Build(TfProviderQuery)
	req.Var("name", name)
	req.Var("vsn", version)
	err := client.Run(req, &resp)
	return resp.TerraformProvider.Content, err
}
