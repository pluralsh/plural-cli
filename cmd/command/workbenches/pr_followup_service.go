package workbenches

import (
	"fmt"
	"strings"
	"time"

	"github.com/pluralsh/plural-cli/pkg/console"
)

const pullRequestNotFoundError = "pull request not found"

type PullRequestURLResolver interface {
	Resolve(options PullRequestOptions) (string, error)
}

type PRFollowupService struct {
	client   console.ConsoleClient
	resolver PullRequestURLResolver
}

type PRFollowupOptions struct {
	Prompt      string
	DeferBy     time.Duration
	PullRequest PullRequestOptions
	SkipMissing bool
}

type PRFollowupResult struct {
	PromptID        string `json:"promptId"`
	PullRequestURL  string `json:"pullRequestUrl"`
	WorkbenchJobURL string `json:"workbenchJobUrl"`
	Skipped         bool   `json:"skipped"`
}

func NewPRFollowupService(client console.ConsoleClient, resolver PullRequestURLResolver) *PRFollowupService {
	return &PRFollowupService{client: client, resolver: resolver}
}

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
