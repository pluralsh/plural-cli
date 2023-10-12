package console

import (
	consoleclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/plural/pkg/api"
)

func (c *consoleClient) ListProviders() (*consoleclient.ListProviders, error) {

	result, err := c.client.ListProviders(c.ctx)
	if err != nil {
		return nil, api.GetErrorResponse(err, "ListProviders")
	}

	return result, nil
}
