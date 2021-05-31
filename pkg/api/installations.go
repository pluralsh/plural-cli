package api

import (
	"fmt"
)

type instResponse struct {
	Installations struct {
		Edges []InstallationEdge
	}
}

var instQuery = fmt.Sprintf(`
	query Installation($name: String!) {
		installation(name: $name) {
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

func (client *Client) GetInstallations() ([]*Installation, error) {
	var resp instResponse
	err := client.Run(client.Build(instsQuery), &resp)
	insts := make([]*Installation, len(resp.Installations.Edges))
	for i, edge := range resp.Installations.Edges {
		insts[i] = edge.Node
	}
	return insts, err
}
