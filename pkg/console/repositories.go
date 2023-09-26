package console

import (
	gqlclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/plural/pkg/api"
)

func (c *consoleClient) CreateRepository(url string, privateKey, passphrase, username, password *string) (*GitRepository, error) {
	attrs := gqlclient.GitAttributes{
		URL:        url,
		PrivateKey: privateKey,
		Passphrase: passphrase,
		Username:   username,
		Password:   password,
	}
	result, err := c.pluralClient.CreateGitRepository(c.ctx, attrs)
	if err != nil {
		return nil, api.GetErrorResponse(err, "CreateGitRepository")
	}

	return convertGitRepository(result.CreateGitRepository), nil
}
