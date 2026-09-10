package agents

import (
	"context"
	"errors"
	"strings"

	gqlclient "github.com/pluralsh/console/go/client"

	pkgagents "github.com/pluralsh/plural-cli/pkg/agents"
	"github.com/pluralsh/plural-cli/pkg/bridge"
)

// Session downloads and restores a Console agent run into a local checkout.
type Session interface {
	Download(context.Context, *gqlclient.AgentRunMinimalFragment) (*pkgagents.SessionBundle, error)
	RestoreAndResume(context.Context, *pkgagents.SessionBundle, string) error
}

func (s *Service) sessions() Session {
	if s.session == nil {
		s.session = pkgagents.NewSessionService(pkgagents.WithSessionInteraction(pkgagents.AcceptingInteraction{}))
	}
	return s.session
}

// Resume downloads the run's session and restores it into repoPath, then
// launches the provider resume command.
func (s *Service) Resume(ctx context.Context, id, repoPath, prRef string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	id = strings.TrimSpace(id)
	repoPath = strings.TrimSpace(repoPath)
	if id == "" {
		return &bridge.Error{Code: bridge.ErrorInvalid, Err: errors.New("agent run id is required")}
	}
	if repoPath == "" {
		return &bridge.Error{Code: bridge.ErrorInvalid, Err: errors.New("local clone path is required")}
	}
	client, err := s.client(ctx)
	if err != nil {
		return err
	}
	run, err := client.GetAgentRun(id)
	if err != nil {
		return err
	}
	if !resumable(run) {
		return &bridge.Error{Code: bridge.ErrorUnavailable, Err: errors.New("agent run has no uploaded session")}
	}
	applyPullRequest(run, prRef)
	bundle, err := s.sessions().Download(ctx, run)
	if err != nil {
		return err
	}
	return s.sessions().RestoreAndResume(ctx, bundle, repoPath)
}

func applyPullRequest(run *gqlclient.AgentRunMinimalFragment, prRef string) {
	prs := pullRequestsWithRef(run)
	if len(prs) == 0 {
		return
	}
	prRef = strings.TrimSpace(prRef)
	if prRef != "" {
		for _, pr := range prs {
			if pr.GetRef() != nil && *pr.GetRef() == prRef {
				run.PullRequests = []*gqlclient.AgentRunMinimalFragment_PullRequests{pr}
				return
			}
		}
	}
	run.PullRequests = []*gqlclient.AgentRunMinimalFragment_PullRequests{prs[0]}
}

func pullRequestsWithRef(run *gqlclient.AgentRunMinimalFragment) []*gqlclient.AgentRunMinimalFragment_PullRequests {
	if run == nil {
		return nil
	}
	prs := make([]*gqlclient.AgentRunMinimalFragment_PullRequests, 0, len(run.GetPullRequests()))
	for _, pr := range run.GetPullRequests() {
		if pr == nil || pr.GetRef() == nil || strings.TrimSpace(*pr.GetRef()) == "" {
			continue
		}
		prs = append(prs, pr)
	}
	return prs
}
