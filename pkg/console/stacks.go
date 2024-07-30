package console

import (
	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/samber/lo"
)

func (c *consoleClient) ListStackRuns(stackID string) (*gqlclient.ListStackRuns, error) {
	return c.client.ListStackRuns(c.ctx, stackID, nil, nil, lo.ToPtr(int64(100)), nil)
}
