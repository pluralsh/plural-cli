package console

import (
	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/samber/lo"
)

func (c *consoleClient) ListStackRuns(stackID string) (*gqlclient.ListStackRuns, error) {
	return c.client.ListStackRuns(c.ctx, stackID, nil, nil, lo.ToPtr(int64(100)), nil)
}

func (c *consoleClient) ListaStacks() (*gqlclient.ListInfrastructureStacks, error) {
	return c.client.ListInfrastructureStacks(c.ctx, nil, lo.ToPtr(int64(100)), nil, nil)
}
