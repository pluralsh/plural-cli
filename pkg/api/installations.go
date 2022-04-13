package api

import (
	"fmt"
)

type instResponse struct {
	Installations struct {
		Edges []InstallationEdge
	}
}

type Binding struct {
	UserId string
	GroupId string
}

type OidcProviderAttributes struct {
	RedirectUris []string
	AuthMethod   string
	Bindings     []Binding
}

var instQuery = fmt.Sprintf(`
	query Installation($name: String, $id: ID) {
		installation(name: $name, id: $id) {
			...InstallationFragment
		}
	}
	%s
`, InstallationFragment)

var instsQuery = fmt.Sprintf(`
	query {
		installations(first: %d) {
			edges { node { ...InstallationFragment } }
		}
	}
	%s
`, pageSize, InstallationFragment)

const oidcProviderMut = `
	mutation OIDCProvider($id: ID!, $attributes: OidcProviderAttributes!) {
		upsertOidcProvider(installationId: $id, attributes: $attributes) {
			id
		}
	}
`

func (client *Client) GetInstallation(name string) (inst *Installation, err error) {
	var resp struct {
		Installation *Installation
	}
	req := client.Build(instQuery)
	req.Var("name", name)
	err = client.Run(req, &resp)
	inst = resp.Installation
	return
}

func (client *Client) GetInstallationById(id string) (inst *Installation, err error) {
	var resp struct {
		Installation *Installation
	}
	req := client.Build(instQuery)
	req.Var("id", id)
	err = client.Run(req, &resp)
	inst = resp.Installation
	return
}

func (client *Client) GetInstallations() ([]*Installation, error) {
	var resp instResponse
	err := client.Run(client.Build(instsQuery), &resp)
	insts := make([]*Installation, len(resp.Installations.Edges))
	for i, edge := range resp.Installations.Edges {
		insts[i] = edge.Node
	}

	fmt.Printf(" resp %s \n  instResponse %s \n", resp, insts)
	return insts, err
}

func (client *Client) OIDCProvider(id string, attributes *OidcProviderAttributes) error {
	var resp struct {
		UpsertOidcProvider struct {
			Id string
		}
	}
	
	req := client.Build(oidcProviderMut)
	req.Var("id", id)
	req.Var("attributes", attributes)
	return client.Run(req, &resp)
}