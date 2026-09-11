// Package workbenches exposes recent Console workbench jobs and queued prompts.
package workbenches

import (
	"context"
	"errors"
	"strings"
	"time"

	gqlclient "github.com/pluralsh/console/go/client"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	"github.com/pluralsh/plural-cli/pkg/console"
)

const defaultPageSize = 10

var errNoConsole = errors.New("connect a Console profile before browsing workbench jobs")

type Summary struct {
	ID            string
	WorkbenchID   string
	WorkbenchName string
	Prompt        string
	Status        string
	InsertedAt    string
}

type Detail struct {
	Summary
	UpdatedAt string
}

type Page struct {
	Items     []Summary
	EndCursor string
	HasNext   bool
}

type PromptResult struct {
	ID          string
	Prompt      string
	DequeueAt   string
	WorkbenchID string
}

type Loader interface {
	List(context.Context, *string, string) (Page, error)
	Get(context.Context, string) (Detail, error)
	FollowUp(context.Context, string, string, time.Duration) (PromptResult, error)
}

type ConsoleResolver interface {
	ActiveConsole(context.Context) (url, token string, err error)
}

type API interface {
	ListWorkbenches(after *string, first *int64, query *string) (*gqlclient.ListWorkbenches_Workbenches, error)
	ListWorkbenchJobs(workbenchID string, page, perPage int) ([]console.WorkbenchJob, error)
	CreateQueuedPrompt(jobID, prompt string, dequeueAt time.Time) (*gqlclient.QueuedPromptFragment, error)
}

type ClientFactory func(token, url string) (API, error)

type Service struct {
	resolve   ConsoleResolver
	newClient ClientFactory
}

func NewService(resolve ConsoleResolver) *Service {
	return &Service{
		resolve: resolve,
		newClient: func(token, url string) (API, error) {
			return console.NewConsoleClient(token, url)
		},
	}
}

func (s *Service) client(ctx context.Context) (API, error) {
	if s.resolve == nil {
		return nil, &bridge.Error{Code: bridge.ErrorUnauthenticated, Err: errNoConsole}
	}
	url, token, err := s.resolve.ActiveConsole(ctx)
	if err != nil {
		return nil, err
	}
	return s.newClient(token, url)
}

func (s *Service) List(ctx context.Context, after *string, query string) (Page, error) {
	if err := ctx.Err(); err != nil {
		return Page{}, err
	}
	client, err := s.client(ctx)
	if err != nil {
		return Page{}, err
	}

	first := int64(20)
	workbenches, err := client.ListWorkbenches(nil, &first, nil)
	if err != nil {
		return Page{}, err
	}
	items := make([]Summary, 0)
	if workbenches != nil {
		for _, edge := range workbenches.GetEdges() {
			if edge == nil || edge.GetNode() == nil {
				continue
			}
			workbench := edge.GetNode()
			jobs, err := client.ListWorkbenchJobs(workbench.GetID(), 1, defaultPageSize)
			if err != nil {
				return Page{}, err
			}
			for _, job := range jobs {
				summary := summaryFromJob(job, workbench.GetName())
				if matches(summary, query) {
					items = append(items, summary)
				}
			}
		}
	}
	return pageItems(items, after, defaultPageSize), nil
}

func (s *Service) Get(ctx context.Context, id string) (Detail, error) {
	if err := ctx.Err(); err != nil {
		return Detail{}, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return Detail{}, &bridge.Error{Code: bridge.ErrorInvalid, Err: errors.New("workbench job id is required")}
	}
	page, err := s.List(ctx, nil, id)
	if err != nil {
		return Detail{}, err
	}
	for _, item := range page.Items {
		if item.ID == id {
			return Detail{Summary: item}, nil
		}
	}
	return Detail{}, &bridge.Error{Code: bridge.ErrorUnavailable, Err: errors.New("workbench job was not found")}
}

func (s *Service) FollowUp(ctx context.Context, jobID, prompt string, deferBy time.Duration) (PromptResult, error) {
	if err := ctx.Err(); err != nil {
		return PromptResult{}, err
	}
	jobID = strings.TrimSpace(jobID)
	prompt = strings.TrimSpace(prompt)
	if jobID == "" {
		return PromptResult{}, &bridge.Error{Code: bridge.ErrorInvalid, Err: errors.New("workbench job id is required")}
	}
	if prompt == "" {
		return PromptResult{}, &bridge.Error{Code: bridge.ErrorInvalid, Err: errors.New("prompt cannot be empty")}
	}
	client, err := s.client(ctx)
	if err != nil {
		return PromptResult{}, err
	}
	dequeueAt := time.Now().Add(deferBy)
	queued, err := client.CreateQueuedPrompt(jobID, prompt, dequeueAt)
	if err != nil {
		return PromptResult{}, err
	}
	return PromptResult{ID: queued.GetID(), Prompt: prompt, DequeueAt: dequeueAt.Format(time.RFC3339Nano), WorkbenchID: jobID}, nil
}

func summaryFromJob(job console.WorkbenchJob, workbenchName string) Summary {
	return Summary{
		ID:            job.ID,
		WorkbenchID:   job.WorkbenchID,
		WorkbenchName: workbenchName,
		Prompt:        job.Prompt,
		Status:        job.Status,
		InsertedAt:    job.InsertedAt,
	}
}

func matches(item Summary, query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return true
	}
	return strings.Contains(strings.ToLower(strings.Join([]string{item.ID, item.WorkbenchID, item.WorkbenchName, item.Prompt, item.Status}, " ")), query)
}

func pageItems(items []Summary, after *string, pageSize int) Page {
	start := 0
	if after != nil {
		for i, item := range items {
			if item.ID == *after {
				start = i + 1
				break
			}
		}
	}
	end := min(len(items), start+pageSize)
	page := Page{Items: items[start:end], HasNext: end < len(items)}
	if len(page.Items) > 0 {
		page.EndCursor = page.Items[len(page.Items)-1].ID
	}
	return page
}
