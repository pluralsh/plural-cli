package console

import (
	consoleclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/plural/pkg/api"
)

func (c *consoleClient) ListClusters() (*consoleclient.ListClusters, error) {

	result, err := c.client.ListClusters(c.ctx, nil, nil, nil)
	if err != nil {
		return nil, api.GetErrorResponse(err, "ListClusters")
	}

	return result, nil
}
