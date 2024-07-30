package console

import (
	"fmt"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/api"
)

func (c *consoleClient) SaveServiceContext(name string, attributes gqlclient.ServiceContextAttributes) (*gqlclient.ServiceContextFragment, error) {
	result, err := c.client.SaveServiceContext(c.ctx, name, attributes)
	if err != nil {
		return nil, api.GetErrorResponse(err, "SaveServiceContext")
	}
	if result == nil {
		return nil, fmt.Errorf("the result from SaveServiceContext is null")
	}
	return result.SaveServiceContext, nil
}

func (c *consoleClient) GetServiceContext(name string) (*gqlclient.ServiceContextFragment, error) {
	result, err := c.client.GetServiceContext(c.ctx, name)
	if err != nil {
		return nil, api.GetErrorResponse(err, "GetServiceContext")
	}
	if result == nil {
		return nil, fmt.Errorf("the result from GetServiceContext is null")
	}
	return result.ServiceContext, nil
}
