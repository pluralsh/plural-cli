// Package agents exposes resumable Console agent runs to presentation layers.
package agents

import (
	"context"
	"errors"
	"strings"

	gqlclient "github.com/pluralsh/console/go/client"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	"github.com/pluralsh/plural-cli/pkg/console"
)

const (
	defaultListLimit int64 = 50
	defaultPageSize        = 10
)

var errNoConsole = errors.New("connect a Console profile before browsing agent runs")

type Summary struct {
	ID         string
	Repository string
	Branch     string
	Provider   string
	Prompt     string
	PRRef      string
}

type Detail struct {
	Summary
	PullRequests []PullRequest
}

type PullRequest struct {
	ID     string
	Ref    string
	Status string
	Title  string
	URL    string
}

type Page struct {
	Items     []Summary
	EndCursor string
	HasNext   bool
}

type Loader interface {
	List(context.Context, *string, string) (Page, error)
	Get(context.Context, string) (Detail, error)
}

type ConsoleResolver interface {
	ActiveConsole(context.Context) (url, token string, err error)
}

type API interface {
	ListAgentRuns(first int64) ([]*gqlclient.AgentRunMinimalFragment, error)
	GetAgentRun(id string) (*gqlclient.AgentRunMinimalFragment, error)
}

type ClientFactory func(token, url string) (API, error)

type Service struct {
	resolve   ConsoleResolver
	newClient ClientFactory
	pageSize  int
}

func NewService(resolve ConsoleResolver) *Service {
	return &Service{
		resolve: resolve,
		newClient: func(token, url string) (API, error) {
			return console.NewConsoleClient(token, url)
		},
		pageSize: defaultPageSize,
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
	runs, err := client.ListAgentRuns(defaultListLimit)
	if err != nil {
		return Page{}, err
	}
	items := make([]Summary, 0, len(runs))
	for _, run := range runs {
		if !resumable(run) {
			continue
		}
		summary := summaryFromRun(run)
		if matches(summary, query) {
			items = append(items, summary)
		}
	}
	return pageItems(items, after, s.pageSize), nil
}

func (s *Service) Get(ctx context.Context, id string) (Detail, error) {
	if err := ctx.Err(); err != nil {
		return Detail{}, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return Detail{}, &bridge.Error{Code: bridge.ErrorInvalid, Err: errors.New("agent run id is required")}
	}
	client, err := s.client(ctx)
	if err != nil {
		return Detail{}, err
	}
	run, err := client.GetAgentRun(id)
	if err != nil {
		return Detail{}, err
	}
	if !resumable(run) {
		return Detail{}, &bridge.Error{Code: bridge.ErrorUnavailable, Err: errors.New("agent run has no uploaded session")}
	}
	return detailFromRun(run), nil
}

func resumable(run *gqlclient.AgentRunMinimalFragment) bool {
	return run != nil && run.GetUpload() != nil && run.GetUpload().GetSession() != nil
}

func summaryFromRun(run *gqlclient.AgentRunMinimalFragment) Summary {
	summary := Summary{ID: run.GetID(), Repository: run.GetRepository(), Prompt: run.GetPrompt()}
	if run.GetBranch() != nil {
		summary.Branch = *run.GetBranch()
	}
	if run.GetRuntime() != nil && run.GetRuntime().GetType() != nil {
		summary.Provider = string(*run.GetRuntime().GetType())
	}
	for _, pr := range run.GetPullRequests() {
		if pr != nil && pr.GetRef() != nil && strings.TrimSpace(*pr.GetRef()) != "" {
			summary.PRRef = *pr.GetRef()
			break
		}
	}
	return summary
}

func detailFromRun(run *gqlclient.AgentRunMinimalFragment) Detail {
	detail := Detail{Summary: summaryFromRun(run)}
	for _, pr := range run.GetPullRequests() {
		if pr == nil || pr.GetRef() == nil || strings.TrimSpace(*pr.GetRef()) == "" {
			continue
		}
		item := PullRequest{ID: pr.GetID(), Ref: *pr.GetRef(), URL: pr.GetURL()}
		if pr.GetStatus() != nil {
			item.Status = string(*pr.GetStatus())
		}
		if pr.GetTitle() != nil {
			item.Title = *pr.GetTitle()
		}
		detail.PullRequests = append(detail.PullRequests, item)
	}
	return detail
}

func matches(item Summary, query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return true
	}
	return strings.Contains(strings.ToLower(strings.Join([]string{item.ID, item.Repository, item.Branch, item.Provider, item.Prompt, item.PRRef}, " ")), query)
}

func pageItems(items []Summary, after *string, pageSize int) Page {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
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
