package workbenches

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePRFollowupClient struct {
	activityID string
	err        error
	url        string
	prompt     string
	calls      int
}

func (f *fakePRFollowupClient) CreateWorkbenchPRFollowup(url, prompt string) (string, error) {
	f.url = url
	f.prompt = prompt
	f.calls++
	return f.activityID, f.err
}

type fakePullRequestResolver struct {
	url     string
	err     error
	options PullRequestOptions
	calls   int
}

func (f *fakePullRequestResolver) Resolve(options PullRequestOptions) (string, error) {
	f.options = options
	f.calls++
	return f.url, f.err
}

func TestPRFollowupServiceUsesResolvedURLAndClientResponse(t *testing.T) {
	client := &fakePRFollowupClient{activityID: "activity-1"}
	resolver := &fakePullRequestResolver{url: "https://github.com/pluralsh/plural-cli/pull/5078"}
	service := NewPRFollowupService(client, resolver)
	options := PRFollowupOptions{
		Prompt:      " verify the fix ",
		PullRequest: PullRequestOptions{Commit: "HEAD~1"},
	}

	result, err := service.Create(options)

	require.NoError(t, err)
	assert.Equal(t, PRFollowupResult{
		ActivityID:     "activity-1",
		PullRequestURL: resolver.url,
	}, result)
	assert.Equal(t, options.PullRequest, resolver.options)
	assert.Equal(t, resolver.url, client.url)
	assert.Equal(t, options.Prompt, client.prompt)
	assert.Equal(t, 1, client.calls)
}

func TestPRFollowupServicePropagatesClientError(t *testing.T) {
	client := &fakePRFollowupClient{err: errors.New("job is currently active")}
	resolver := &fakePullRequestResolver{url: "https://github.com/pluralsh/plural-cli/pull/5078"}
	service := NewPRFollowupService(client, resolver)

	_, err := service.Create(PRFollowupOptions{Prompt: "verify the fix"})

	require.EqualError(t, err, "job is currently active")
}

func TestPRFollowupServiceSkipsMissingPullRequest(t *testing.T) {
	client := &fakePRFollowupClient{err: errors.New("GraphQL error: pull request not found: WorkbenchPrFollowup")}
	resolver := &fakePullRequestResolver{url: "https://github.com/pluralsh/plural-cli/pull/5078"}
	service := NewPRFollowupService(client, resolver)

	result, err := service.Create(PRFollowupOptions{
		Prompt:      "verify the fix",
		SkipMissing: true,
	})

	require.NoError(t, err)
	assert.Equal(t, PRFollowupResult{
		PullRequestURL: resolver.url,
		Skipped:        true,
	}, result)
	assert.Equal(t, 1, client.calls)
}

func TestPRFollowupServiceDoesNotSkipOtherErrors(t *testing.T) {
	client := &fakePRFollowupClient{err: errors.New("GraphQL error: unauthorized")}
	resolver := &fakePullRequestResolver{url: "https://github.com/pluralsh/plural-cli/pull/5078"}
	service := NewPRFollowupService(client, resolver)

	_, err := service.Create(PRFollowupOptions{
		Prompt:      "verify the fix",
		SkipMissing: true,
	})

	require.EqualError(t, err, "GraphQL error: unauthorized")
}

func TestPRFollowupServiceDoesNotSkipUnrelatedErrorContainingMissingText(t *testing.T) {
	client := &fakePRFollowupClient{err: errors.New("failed while checking whether pull request not found should be ignored")}
	resolver := &fakePullRequestResolver{url: "https://github.com/pluralsh/plural-cli/pull/5078"}
	service := NewPRFollowupService(client, resolver)

	_, err := service.Create(PRFollowupOptions{
		Prompt:      "verify the fix",
		SkipMissing: true,
	})

	require.EqualError(t, err, "failed while checking whether pull request not found should be ignored")
}

func TestPRFollowupServiceRejectsEmptyClientResponse(t *testing.T) {
	client := &fakePRFollowupClient{}
	resolver := &fakePullRequestResolver{url: "https://github.com/pluralsh/plural-cli/pull/5078"}
	service := NewPRFollowupService(client, resolver)

	_, err := service.Create(PRFollowupOptions{Prompt: "verify the fix"})

	require.EqualError(t, err, "console returned an empty workbench PR follow-up response")
}

func TestPRFollowupServiceValidatesPromptBeforeResolvingURL(t *testing.T) {
	client := &fakePRFollowupClient{}
	resolver := &fakePullRequestResolver{err: errors.New("should not be called")}
	service := NewPRFollowupService(client, resolver)

	_, err := service.Create(PRFollowupOptions{Prompt: " \t "})

	require.EqualError(t, err, "prompt cannot be empty")
	assert.Zero(t, resolver.calls)
	assert.Zero(t, client.calls)
}
