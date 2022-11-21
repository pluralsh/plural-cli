package api

import (
	"github.com/pluralsh/gqlclient"
)

func (client *client) DestroyCluster(domain, name, provider string) error {
	_, err := client.pluralClient.DestroyCluster(client.ctx, domain, name, gqlclient.Provider(NormalizeProvider(provider)))
	if err != nil {
		return err
	}

	return nil
}
