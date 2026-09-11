package console

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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

func (c *consoleClient) ListWorkbenches(after *string, first *int64, query *string) (*consoleclient.ListWorkbenches_Workbenches, error) {
	result, err := c.client.ListWorkbenches(c.ctx, after, first, nil, nil, query)
	if err != nil {
		return nil, api.GetErrorResponse(err, "ListWorkbenches")
	}
	return result.GetWorkbenches(), nil
}

func (c *consoleClient) GetWorkbench(id string) (*consoleclient.WorkbenchFragment, error) {
	result, err := c.client.GetWorkbench(c.ctx, &id, nil)
	if err != nil {
		return nil, api.GetErrorResponse(err, "GetWorkbench")
	}
	return result.GetWorkbench(), nil
}

func (c *consoleClient) ListWorkbenchJobs(workbenchID string, page, perPage int) ([]WorkbenchJob, error) {
	endpoint, err := url.Parse(c.url)
	if err != nil {
		return nil, err
	}
	endpoint.Path = fmt.Sprintf("/v1/api/ai/workbenches/%s/jobs", workbenchID)
	query := endpoint.Query()
	query.Set("page", fmt.Sprintf("%d", page))
	query.Set("per_page", fmt.Sprintf("%d", perPage))
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(c.ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Token "+c.token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("list workbench jobs: %s", resp.Status)
	}
	var result struct {
		Data []WorkbenchJob `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func (c *consoleClient) CreateQueuedPrompt(jobID, prompt string, dequeueAt time.Time) (*consoleclient.QueuedPromptFragment, error) {
	result, err := c.client.CreateQueuedPrompt(c.ctx, jobID, consoleclient.QueuedPromptAttributes{
		Prompt:      prompt,
		DequeableAt: dequeueAt.Format(time.RFC3339Nano),
	})
	if err != nil {
		return nil, api.GetErrorResponse(err, "CreateQueuedPrompt")
	}
	fragment := result.GetCreateQueuedPrompt()
	if fragment == nil {
		return nil, fmt.Errorf("returned object [CreateQueuedPrompt] is nil")
	}
	return fragment, nil
}
