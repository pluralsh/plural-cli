package api

import (
	"github.com/pluralsh/gqlclient"
	"sigs.k8s.io/yaml"
)

func (client *client) CreateUpgrade(queue, repository string, attrs gqlclient.UpgradeAttributes) error {
	_, err := client.pluralClient.CreateUpgrade(client.ctx, queue, repository, attrs)
	return err
}

func ConstructUpgradeAttributes(marshalled []byte) (gqlclient.UpgradeAttributes, error) {
	var attrs gqlclient.UpgradeAttributes
	err := yaml.Unmarshal(marshalled, &attrs)
	return attrs, err
}
