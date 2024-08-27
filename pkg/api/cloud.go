package api

import (
	"github.com/pluralsh/gqlclient"
)

func (client *client) GetConsoleInstances() ([]*gqlclient.ConsoleInstanceFragment, error) {
	res := []*gqlclient.ConsoleInstanceFragment{}
	resp, err := client.pluralClient.GetConsoleInstances(client.ctx, 100)
	if err != nil {
		return res, err
	}

	for _, node := range resp.ConsoleInstances.Edges {
		res = append(res, node.Node)
	}

	return res, nil
}

func (client *client) UpdateConsoleInstance(id string, attrs gqlclient.ConsoleInstanceUpdateAttributes) error {
	_, err := client.pluralClient.UpdateConsoleInstance(client.ctx, id, attrs)
	return err
}
