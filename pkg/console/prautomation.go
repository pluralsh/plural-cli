package console

import (
	gqlclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/plural-cli/pkg/api"
)

func (c *consoleClient) CreatePullRequest(id string, name, branch, context *string) (*gqlclient.PullRequestFragment, error) {
	result, err := c.client.CreatePullRequest(c.ctx, id, branch, context)
	if err != nil {
		return nil, api.GetErrorResponse(err, "CreatePullRequest")
	}

	return result.CreatePullRequest, nil
}
