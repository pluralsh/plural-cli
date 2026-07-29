package workbenches

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePullRequestRepository struct {
	subject        string
	remote         string
	subjectErr     error
	remoteErr      error
	requestedRef   string
	subjectCalls   int
	remoteURLCalls int
}

func (f *fakePullRequestRepository) CommitSubject(ref string) (string, error) {
	f.requestedRef = ref
	f.subjectCalls++
	return f.subject, f.subjectErr
}

func (f *fakePullRequestRepository) RemoteURL() (string, error) {
	f.remoteURLCalls++
	return f.remote, f.remoteErr
}

func TestPullRequestResolverUsesExplicitURL(t *testing.T) {
	repository := &fakePullRequestRepository{subjectErr: errors.New("should not be called")}
	resolver := NewPullRequestResolver(repository)

	result, err := resolver.Resolve(PullRequestOptions{URL: "https://github.com/pluralsh/console/pull/3905"})

	require.NoError(t, err)
	assert.Equal(t, "https://github.com/pluralsh/console/pull/3905", result)
	assert.Zero(t, repository.subjectCalls)
	assert.Zero(t, repository.remoteURLCalls)
}

func TestPullRequestResolverDoesNotTrimExplicitURL(t *testing.T) {
	resolver := NewPullRequestResolver(&fakePullRequestRepository{})

	_, err := resolver.Resolve(PullRequestOptions{URL: " https://github.com/pluralsh/console/pull/3905 "})

	require.ErrorContains(t, err, "invalid pull request URL")
}

func TestPullRequestResolverRejectsURLAndCommit(t *testing.T) {
	resolver := NewPullRequestResolver(&fakePullRequestRepository{})

	_, err := resolver.Resolve(PullRequestOptions{
		URL:    "https://github.com/pluralsh/console/pull/3905",
		Commit: "HEAD~1",
	})

	require.EqualError(t, err, "url and commit cannot be used together")
}

func TestPullRequestResolverInfersProviderURL(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		remote   string
		expected string
	}{
		{
			name:     "GitHub squash commit over SSH",
			subject:  "Implement verification (#5078)",
			remote:   "git@github.com:pluralsh/plural-cli.git",
			expected: "https://github.com/pluralsh/plural-cli/pull/5078",
		},
		{
			name:     "GitHub merge commit over HTTPS",
			subject:  "Merge pull request #3905 from pluralsh/more-perf-improvements",
			remote:   "https://github.com/pluralsh/console.git",
			expected: "https://github.com/pluralsh/console/pull/3905",
		},
		{
			name:     "GitLab merge request over SSH URL",
			subject:  "See merge request group/project!73",
			remote:   "ssh://git@gitlab.com/group/project.git",
			expected: "https://gitlab.com/group/project/-/merge_requests/73",
		},
		{
			name:     "Bitbucket pull request",
			subject:  "Merge pull request #19",
			remote:   "https://bitbucket.org/team/project.git",
			expected: "https://bitbucket.org/team/project/pull-requests/19",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &fakePullRequestRepository{subject: tt.subject, remote: tt.remote}
			resolver := NewPullRequestResolver(repository)

			result, err := resolver.Resolve(PullRequestOptions{})

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, "HEAD", repository.requestedRef)
			assert.Equal(t, 1, repository.remoteURLCalls)
		})
	}
}

func TestPullRequestResolverUsesExplicitProviderAndBaseURL(t *testing.T) {
	repository := &fakePullRequestRepository{subject: "See merge request group/project!81"}
	resolver := NewPullRequestResolver(repository)

	result, err := resolver.Resolve(PullRequestOptions{
		Commit:   "HEAD~1",
		BaseURL:  "https://code.example.com/team/project",
		Provider: "gitlab",
	})

	require.NoError(t, err)
	assert.Equal(t, "https://code.example.com/team/project/-/merge_requests/81", result)
	assert.Zero(t, repository.remoteURLCalls)
}

func TestPullRequestResolverUsesExplicitProviderWithSelfHostedOrigin(t *testing.T) {
	repository := &fakePullRequestRepository{
		subject: "See merge request group/project!82",
		remote:  "git@code.example.com:group/project.git",
	}
	resolver := NewPullRequestResolver(repository)

	result, err := resolver.Resolve(PullRequestOptions{Provider: "gitlab"})

	require.NoError(t, err)
	assert.Equal(t, "https://code.example.com/group/project/-/merge_requests/82", result)
	assert.Equal(t, 1, repository.remoteURLCalls)
}

func TestPullRequestResolverUsesProviderSpecificCommitPattern(t *testing.T) {
	repository := &fakePullRequestRepository{
		subject: "GitHub-style title (#82)",
		remote:  "git@code.example.com:group/project.git",
	}
	resolver := NewPullRequestResolver(repository)

	_, err := resolver.Resolve(PullRequestOptions{Provider: "gitlab"})

	require.EqualError(t, err, `commit "HEAD" does not identify a gitlab pull request`)
}

func TestPullRequestResolverRejectsUnsupportedProvider(t *testing.T) {
	resolver := NewPullRequestResolver(&fakePullRequestRepository{})

	_, err := resolver.Resolve(PullRequestOptions{Provider: "azure-devops"})

	require.EqualError(t, err, `unsupported source control provider "azure-devops"`)
}

func TestPullRequestResolverErrors(t *testing.T) {
	tests := []struct {
		name       string
		subject    string
		remote     string
		subjectErr error
		remoteErr  error
		errorText  string
	}{
		{
			name:       "commit lookup fails",
			subjectErr: errors.New("unknown revision"),
			errorText:  `could not read commit "HEAD": unknown revision`,
		},
		{
			name:      "subject has no pull request",
			subject:   "ordinary commit",
			remote:    "git@github.com:team/project.git",
			errorText: `commit "HEAD" does not identify a github pull request`,
		},
		{
			name:      "remote lookup fails",
			subject:   "Fix issue (#12)",
			remoteErr: errors.New("origin missing"),
			errorText: "could not read origin URL: origin missing",
		},
		{
			name:      "unknown provider",
			subject:   "Fix issue (#12)",
			remote:    "git@code.example.com:team/project.git",
			errorText: `cannot infer source control provider from host "code.example.com"; provide --provider`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &fakePullRequestRepository{
				subject:    tt.subject,
				remote:     tt.remote,
				subjectErr: tt.subjectErr,
				remoteErr:  tt.remoteErr,
			}
			resolver := NewPullRequestResolver(repository)

			_, err := resolver.Resolve(PullRequestOptions{})

			require.EqualError(t, err, tt.errorText)
		})
	}
}

func TestPullRequestResolverParsesRepositoryAddress(t *testing.T) {
	tests := []struct {
		raw      string
		expected repositoryAddress
	}{
		{
			raw:      "git@github.com:pluralsh/plural-cli.git",
			expected: repositoryAddress{url: "https://github.com/pluralsh/plural-cli", host: "github.com"},
		},
		{
			raw:      "ssh://git@gitlab.example.com/group/project.git",
			expected: repositoryAddress{url: "https://gitlab.example.com/group/project", host: "gitlab.example.com"},
		},
		{
			raw:      "https://bitbucket.org/team/project.git/",
			expected: repositoryAddress{url: "https://bitbucket.org/team/project", host: "bitbucket.org"},
		},
	}
	resolver := NewPullRequestResolver(&fakePullRequestRepository{})

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			address, err := resolver.parseRepositoryAddress(tt.raw)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, address)
		})
	}
}
