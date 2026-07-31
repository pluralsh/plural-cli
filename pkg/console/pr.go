package console

import (
	"fmt"

	consoleclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/api"
)

func (c *consoleClient) CreatePullRequest(id string, branch, context *string) (*consoleclient.PullRequestFragment, error) {
	result, err := c.client.CreatePullRequest(c.ctx, id, nil, branch, context)
	if err != nil {
		return nil, api.GetErrorResponse(err, "CreatePullRequest")
	}

	return result.CreatePullRequest, nil
}

func (c *consoleClient) ListPrAutomations() (*consoleclient.ListPrAutomations, error) {
	result, err := c.client.ListPrAutomations(c.ctx, nil, nil, nil)
	if err != nil {
		return nil, api.GetErrorResponse(err, "ListPrAutomations")
	}
	return result, nil
}

func (c *consoleClient) GetPrAutomation(id string) (*consoleclient.PrAutomationFragment, error) {
	result, err := c.client.GetPrAutomation(c.ctx, id)
	if err != nil {
		return nil, api.GetErrorResponse(err, "GetPrAutomation")
	}
	if result == nil || result.PrAutomation == nil {
		return nil, fmt.Errorf("pr automation %s was not found", id)
	}
	return result.PrAutomation, nil
}

func (c *consoleClient) GetPrAutomationByName(name string) (*consoleclient.PrAutomationFragment, error) {
	result, err := c.client.GetPrAutomationByName(c.ctx, name)
	if err != nil {
		return nil, api.GetErrorResponse(err, "GetPrAutomationByName")
	}

	if result.PrAutomation == nil {
		return nil, fmt.Errorf("pr automation %s not found", name)
	}

	return result.PrAutomation, nil
}
