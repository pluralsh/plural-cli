package workbenches

import (
	"regexp"
	"strings"
)

type ProviderName string

const (
	ProviderAuto      ProviderName = "auto"
	ProviderGitHub    ProviderName = "github"
	ProviderGitLab    ProviderName = "gitlab"
	ProviderBitbucket ProviderName = "bitbucket"
)

type PullRequestProvider interface {
	Name() ProviderName
	Supports(host string) bool
	PullRequestNumber(subject string) (string, bool)
	PullRequestURL(repositoryURL, number string) string
}

type pullRequestProvider struct {
	name       ProviderName
	hostMarker string
	path       string
	patterns   []*regexp.Regexp
}

func defaultPullRequestProviders() []PullRequestProvider {
	return []PullRequestProvider{
		pullRequestProvider{
			name:       ProviderGitHub,
			hostMarker: "github",
			path:       "/pull/",
			patterns: []*regexp.Regexp{
				regexp.MustCompile(`\(#([0-9]+)\)`),
				regexp.MustCompile(`(?i)merge pull request #([0-9]+)\b`),
			},
		},
		pullRequestProvider{
			name:       ProviderGitLab,
			hostMarker: "gitlab",
			path:       "/-/merge_requests/",
			patterns: []*regexp.Regexp{
				regexp.MustCompile(`!([0-9]+)\b`),
			},
		},
		pullRequestProvider{
			name:       ProviderBitbucket,
			hostMarker: "bitbucket",
			path:       "/pull-requests/",
			patterns: []*regexp.Regexp{
				regexp.MustCompile(`(?i)pull request #([0-9]+)\b`),
			},
		},
	}
}

func (p pullRequestProvider) Name() ProviderName {
	return p.name
}

func (p pullRequestProvider) Supports(host string) bool {
	return strings.Contains(strings.ToLower(host), p.hostMarker)
}

func (p pullRequestProvider) PullRequestNumber(subject string) (string, bool) {
	for _, pattern := range p.patterns {
		matches := pattern.FindStringSubmatch(subject)

		if len(matches) == 2 {
			return matches[1], true
		}
	}

	return "", false
}

func (p pullRequestProvider) PullRequestURL(repositoryURL, number string) string {
	return repositoryURL + p.path + number
}
