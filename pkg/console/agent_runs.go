package console

import (
	console "github.com/pluralsh/console/go/client"
	"github.com/samber/lo"

	"github.com/pluralsh/plural-cli/pkg/api"
)

func (c *consoleClient) GetAgentRun(id string) (*console.AgentRunMinimalFragment, error) {
	result, err := c.client.GetAgentRunMinimal(c.ctx, id)
	if err != nil {
		return nil, api.GetErrorResponse(err, "GetAgentRun")
	}

	return result.GetAgentRun(), nil
}

func (c *consoleClient) ListAgentRuns(first int64) ([]*console.AgentRunMinimalFragment, error) {
	result, err := c.client.ListAgentRunsMinimal(c.ctx, nil, new(first), nil, nil)
	if err != nil {
		return nil, api.GetErrorResponse(err, "ListAgentRuns")
	}

	return lo.Map(result.GetAgentRuns().GetEdges(), func(item *console.ListAgentRunsMinimal_AgentRuns_Edges, index int) *console.AgentRunMinimalFragment {
		return item.GetNode()
	}), nil
}
