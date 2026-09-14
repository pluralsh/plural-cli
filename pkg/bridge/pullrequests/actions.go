package pullrequests

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/samber/lo"

	"github.com/pluralsh/plural-cli/pkg/bridge"
)

var (
	errMissingAutomationID = errors.New("pr automation id is required")
	errMissingPR           = errors.New("pull request was not created")
)

// CreatePRInput creates a PR from an automation id (CLI: plural pr create).
type CreatePRInput struct {
	AutomationID string
	Branch       string
	Context      string // optional raw JSON
}

// TriggerPRInput triggers an automation with configuration (CLI: plural pr trigger).
type TriggerPRInput struct {
	AutomationID  string
	Name          string
	Branch        string
	Configuration map[string]string
}

// CreatedPR is the credential-free result of create/trigger.
type CreatedPR struct {
	ID      string
	URL     string
	Title   string
	Status  string
	Creator string
	Ref     string
}

// Loader is the narrow contract consumed by the Pull requests screen.
type Loader interface {
	List(ctx context.Context, after *string, query string) (Page, error)
	Get(ctx context.Context, id string) (Detail, error)
	CreatePR(ctx context.Context, input CreatePRInput) (CreatedPR, error)
	TriggerPR(ctx context.Context, input TriggerPRInput) (CreatedPR, error)
}

// API is the Console surface required by this package.
type API interface {
	ListPrAutomations() (*gqlclient.ListPrAutomations, error)
	GetPrAutomation(id string) (*gqlclient.PrAutomationFragment, error)
	CreatePullRequest(id string, branch, context *string) (*gqlclient.PullRequestFragment, error)
}

func (s *Service) CreatePR(ctx context.Context, input CreatePRInput) (CreatedPR, error) {
	if err := ctx.Err(); err != nil {
		return CreatedPR{}, err
	}
	id := strings.TrimSpace(input.AutomationID)
	if id == "" {
		return CreatedPR{}, &bridge.Error{Code: bridge.ErrorInvalid, Err: errMissingAutomationID}
	}
	client, err := s.client(ctx)
	if err != nil {
		return CreatedPR{}, err
	}
	var branch, context *string
	if b := strings.TrimSpace(input.Branch); b != "" {
		branch = &b
	}
	if c := strings.TrimSpace(input.Context); c != "" {
		context = &c
	}
	pr, err := client.CreatePullRequest(id, branch, context)
	if err != nil {
		return CreatedPR{}, err
	}
	if pr == nil {
		return CreatedPR{}, &bridge.Error{Code: bridge.ErrorUnavailable, Err: errMissingPR}
	}
	return createdFromFragment(pr), nil
}

func (s *Service) TriggerPR(ctx context.Context, input TriggerPRInput) (CreatedPR, error) {
	if err := ctx.Err(); err != nil {
		return CreatedPR{}, err
	}
	id := strings.TrimSpace(input.AutomationID)
	if id == "" {
		return CreatedPR{}, &bridge.Error{Code: bridge.ErrorInvalid, Err: errMissingAutomationID}
	}
	client, err := s.client(ctx)
	if err != nil {
		return CreatedPR{}, err
	}
	cfg := input.Configuration
	if cfg == nil {
		cfg = map[string]string{}
	}
	contextJSON, err := json.Marshal(cfg)
	if err != nil {
		return CreatedPR{}, err
	}
	var branch *string
	if b := strings.TrimSpace(input.Branch); b != "" {
		branch = &b
	}
	pr, err := client.CreatePullRequest(id, branch, lo.ToPtr(string(contextJSON)))
	if err != nil {
		return CreatedPR{}, err
	}
	if pr == nil {
		return CreatedPR{}, &bridge.Error{Code: bridge.ErrorUnavailable, Err: errMissingPR}
	}
	return createdFromFragment(pr), nil
}

func createdFromFragment(pr *gqlclient.PullRequestFragment) CreatedPR {
	created := CreatedPR{ID: pr.ID, URL: pr.URL}
	if pr.Title != nil {
		created.Title = *pr.Title
	}
	if pr.Status != nil {
		created.Status = string(*pr.Status)
	}
	if pr.Creator != nil {
		created.Creator = *pr.Creator
	}
	if pr.Ref != nil {
		created.Ref = *pr.Ref
	}
	return created
}
