package console

import (
	"fmt"
	"time"

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

func (c *consoleClient) EnqueueWorkbenchPRFollowup(url, prompt string, dur time.Duration) (string, error) {
	dequeableAt := time.Now().Add(dur)
	result, err := c.client.EnqueueWorkbenchPrFollowup(c.ctx, url, consoleclient.QueuedPromptAttributes{
		Prompt:      prompt,
		DequeableAt: dequeableAt.Format(time.RFC3339),
	})
	if err != nil {
		return "", api.GetErrorResponse(err, "EnqueueWorkbenchPrFollowup")
	}
	return result.GetEnqueueWorkbenchPrFollowup().GetID(), nil
}
