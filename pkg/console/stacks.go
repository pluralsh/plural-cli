package console

import (
	"fmt"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/samber/lo"

	"github.com/pluralsh/plural-cli/pkg/api"
)

func (c *consoleClient) ListStackRuns(stackID string) (*gqlclient.ListStackRuns, error) {
	result, err := c.client.ListStackRuns(c.ctx, stackID, nil, nil, lo.ToPtr(int64(100)), nil)
	if err != nil {
		return nil, api.GetErrorResponse(err, "ListStackRuns")
	}
	return result, nil
}

func (c *consoleClient) ListStacks() (*gqlclient.ListInfrastructureStacks, error) {
	result, err := c.client.ListInfrastructureStacks(c.ctx, nil, lo.ToPtr(int64(100)), nil, nil)
	if err != nil {
		return nil, api.GetErrorResponse(err, "ListInfrastructureStacks")
	}
	return result, nil
}

func (c *consoleClient) GetStack(id string) (*gqlclient.InfrastructureStackFragment, error) {
	response, err := c.client.GetInfrastructureStack(c.ctx, lo.ToPtr(id), nil)
	if err != nil {
		return nil, api.GetErrorResponse(err, "GetInfrastructureStack")
	}
	if response == nil || response.InfrastructureStack == nil {
		return nil, fmt.Errorf("infrastructure stack %s was not found", id)
	}
	return response.InfrastructureStack, nil
}
