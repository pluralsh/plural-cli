package console

import (
	gqlclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/plural/pkg/api"
)

func (c *consoleClient) ListClusterServices(clusterId string) (*gqlclient.ListServiceDeployment, error) {
	result, err := c.client.ListServiceDeployment(c.ctx, nil, nil, nil, &clusterId)
	if err != nil {
		return nil, api.GetErrorResponse(err, "ListClusterServices")
	}

	return result, nil
}
