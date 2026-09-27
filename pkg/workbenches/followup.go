// Package workbenches implements Console workbench PR follow-up, shared by CLI and TUI.
package workbenches

import (
	"fmt"
	"strings"
	"time"

	consoleclient "github.com/pluralsh/console/go/client"
)

const pullRequestNotFoundError = "pull request not found"

// Enqueuer queues a follow-up prompt against the workbench job for a pull request.
type Enqueuer interface {
	EnqueueWorkbenchPRFollowup(url, prompt string, deferBy time.Duration) (*consoleclient.EnqueueWorkbenchPrFollowup_EnqueueWorkbenchPrFollowup, error)
}

// PullRequestURLResolver turns CLI pull-request flags into a canonical URL.
type PullRequestURLResolver interface {
	Resolve(options PullRequestOptions) (string, error)
}

// PRFollowupService queues a follow-up prompt for the workbench job on a PR.
type PRFollowupService struct {
	client   Enqueuer
	resolver PullRequestURLResolver
}

// PRFollowupOptions are the inputs for Create.
type PRFollowupOptions struct {
	Prompt      string
	DeferBy     time.Duration
	PullRequest PullRequestOptions
	SkipMissing bool
}

// PRFollowupResult is the queued follow-up returned by Console.
type PRFollowupResult struct {
	PromptID        string `json:"promptId"`
	PullRequestURL  string `json:"pullRequestUrl"`
	WorkbenchJobURL string `json:"workbenchJobUrl"`
	Skipped         bool   `json:"skipped"`
}

// NewPRFollowupService constructs a follow-up service.
func NewPRFollowupService(client Enqueuer, resolver PullRequestURLResolver) *PRFollowupService {
	return &PRFollowupService{client: client, resolver: resolver}
}

// Create resolves the pull request and queues the follow-up prompt.
func (s *PRFollowupService) Create(options PRFollowupOptions) (PRFollowupResult, error) {
	if strings.TrimSpace(options.Prompt) == "" {
		return PRFollowupResult{}, fmt.Errorf("prompt cannot be empty")
	}
	if s.client == nil {
		return PRFollowupResult{}, fmt.Errorf("workbench PR follow-up client is not configured")
	}
	if s.resolver == nil {
		return PRFollowupResult{}, fmt.Errorf("pull request URL resolver is not configured")
	}

	pullRequestURL, err := s.resolver.Resolve(options.PullRequest)
	if err != nil {
		return PRFollowupResult{}, err
	}

	result, err := s.client.EnqueueWorkbenchPRFollowup(pullRequestURL, options.Prompt, options.DeferBy)
	if err != nil {
		if options.SkipMissing && s.isPullRequestNotFound(err) {
			return PRFollowupResult{PullRequestURL: pullRequestURL, Skipped: true}, nil
		}

		return PRFollowupResult{}, err
	}
	if result == nil {
		return PRFollowupResult{}, fmt.Errorf("console returned an empty workbench PR follow-up response")
	}

	promptID := result.GetID()
	if promptID == "" {
		return PRFollowupResult{}, fmt.Errorf("console returned an empty workbench PR follow-up response")
	}

	workbenchJobURL := result.GetWorkbenchJob().GetURL()
	if workbenchJobURL == "" {
		return PRFollowupResult{}, fmt.Errorf("console returned an empty workbench job URL")
	}

	return PRFollowupResult{
		PromptID:        promptID,
		PullRequestURL:  pullRequestURL,
		WorkbenchJobURL: workbenchJobURL,
	}, nil
}

func (s *PRFollowupService) isPullRequestNotFound(err error) bool {
	if err == nil {
		return false
	}

	message := err.Error()
	return message == pullRequestNotFoundError ||
		strings.HasPrefix(message, pullRequestNotFoundError+":") ||
		strings.HasSuffix(message, ": "+pullRequestNotFoundError) ||
		strings.Contains(message, ": "+pullRequestNotFoundError+":")
}
