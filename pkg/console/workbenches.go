package console

import (
	"fmt"

	consoleclient "github.com/pluralsh/console/go/client"

	"github.com/pluralsh/plural-cli/pkg/api"
)

func (c *consoleClient) CreateWorkbenchPRFollowup(url, prompt string) (string, error) {
	result, err := c.client.WorkbenchPrFollowup(c.ctx, url, consoleclient.WorkbenchMessageAttributes{Prompt: prompt})
	if err != nil {
		return "", api.GetErrorResponse(err, "WorkbenchPrFollowup")
	}
	activity := result.GetWorkbenchPrFollowup()
	if activity == nil {
		return "", fmt.Errorf("returned object [WorkbenchPrFollowup] is nil")
	}

	return activity.GetID(), nil
}
