package console

import (
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
