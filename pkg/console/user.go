package console

import (
	consoleclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/api"
)

func (c *consoleClient) GetUser(email string) (*consoleclient.UserFragment, error) {
	response, err := c.client.GetUser(c.ctx, email)
	if err != nil {
		return nil, api.GetErrorResponse(err, "GetUser")
	}

	return response.User, err
}
