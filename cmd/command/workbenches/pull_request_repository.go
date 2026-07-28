package workbenches

import gitutils "github.com/pluralsh/plural-cli/pkg/utils/git"

type PullRequestRepository interface {
	CommitSubject(ref string) (string, error)
	RemoteURL() (string, error)
}

type GitPullRequestRepository struct{}

func (GitPullRequestRepository) CommitSubject(ref string) (string, error) {
	return gitutils.GitRaw("show", "-s", "--format=%s", ref)
}

func (GitPullRequestRepository) RemoteURL() (string, error) {
	return gitutils.GitRaw("remote", "get-url", "origin")
}
