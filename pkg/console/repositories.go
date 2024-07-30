package console

import (
	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/api"
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

func (c *consoleClient) UpdateRepository(id string, attrs gqlclient.GitAttributes) (*gqlclient.UpdateGitRepository, error) {

	res, err := c.client.UpdateGitRepository(c.ctx, id, attrs)
	if err != nil {
		return nil, api.GetErrorResponse(err, "UpdateGitRepository")
	}
	return res, nil
}
