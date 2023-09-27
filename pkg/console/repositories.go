package console

import (
	gqlclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/plural/pkg/api"
)

func (c *consoleClient) CreateRepository(url string, privateKey, passphrase, username, password *string) (*gqlclient.CreateGitRepository, error) {
	attrs := gqlclient.GitAttributes{
		URL:        url,
		PrivateKey: privateKey,
		Passphrase: passphrase,
		Username:   username,
		Password:   password,
	}
	res, err := c.client.CreateGitRepository(c.ctx, attrs)
	if err != nil {
		return nil, api.GetErrorResponse(err, "CreateGitRepository")
	}
	return res, nil
}

func (c *consoleClient) ListRepositories() (*gqlclient.ListGitRepositories, error) {
	result, err := c.client.ListGitRepositories(c.ctx, nil, nil, nil)
	if err != nil {
		return nil, api.GetErrorResponse(err, "ListRepositories")
	}

	return result, nil
}
