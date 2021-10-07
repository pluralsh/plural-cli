package api

import (
	"fmt"
)

var provisionDomain = fmt.Sprintf(`
	mutation Create($name: String!) {
		provisionDomain(name: $name) {
			...DnsDomainFragment
		}
	}
	%s
`, DnsDomainFragment)

func (client *Client) CreateDomain(name string) error {
	var resp struct {
		ProvisionDomain *DnsDomain
	}

	req := client.Build(provisionDomain)
	req.Var("name", name)
	return client.Run(req, &resp)
}
