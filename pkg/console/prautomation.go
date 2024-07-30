package console

import (
	gqlclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/plural-cli/pkg/api"
)

func (c *consoleClient) CreatePrAutomation(attrs gqlclient.PrAutomationAttributes) (*gqlclient.PrAutomationFragment, error) {
	result, err := c.client.CreatePrAutomation(c.ctx, attrs)
	if err != nil {
		return nil, api.GetErrorResponse(err, "CreatePrAutomation")
	}

	return result.CreatePrAutomation, nil
}
