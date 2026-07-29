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

func (c *consoleClient) EnqueueWorkbenchPRFollowup(url, prompt string, deferBy time.Duration) (*consoleclient.EnqueueWorkbenchPrFollowup_EnqueueWorkbenchPrFollowup, error) {
	dequeueAt := time.Now().Add(deferBy)
	result, err := c.client.EnqueueWorkbenchPrFollowup(c.ctx, url, consoleclient.QueuedPromptAttributes{
		Prompt:      prompt,
		DequeableAt: dequeueAt.Format(time.RFC3339Nano),
	})
	if err != nil {
		return nil, api.GetErrorResponse(err, "EnqueueWorkbenchPrFollowup")
	}
	fragment := result.GetEnqueueWorkbenchPrFollowup()
	if fragment == nil {
		return nil, fmt.Errorf("returned object [EnqueueWorkbenchPrFollowup] is nil")
	}

	return fragment, nil
}
