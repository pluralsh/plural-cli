package agents

import (
	"context"
	"errors"
	"testing"

	gqlclient "github.com/pluralsh/console/go/client"

	pkgagents "github.com/pluralsh/plural-cli/pkg/agents"
	"github.com/pluralsh/plural-cli/pkg/bridge"
)

type fakeSession struct {
	runID string
	path  string
	err   error
}

func (f *fakeSession) Download(_ context.Context, run *gqlclient.AgentRunMinimalFragment) (*pkgagents.SessionBundle, error) {
	if run != nil {
		f.runID = run.ID
	}
	if f.err != nil {
		return nil, f.err
	}
	return &pkgagents.SessionBundle{Run: run, Manifest: &pkgagents.SessionManifest{}}, nil
}

func (f *fakeSession) RestoreAndResume(_ context.Context, _ *pkgagents.SessionBundle, path string) error {
	f.path = path
	return f.err
}

func resumableRun(id, prRef string) *gqlclient.AgentRunMinimalFragment {
	session := "https://example.com/session.tgz"
	provider := gqlclient.AgentRuntimeTypeCodex
	branch := "main"
	run := &gqlclient.AgentRunMinimalFragment{
		ID:         id,
		Repository: "git@github.com:acme/repo.git",
		Branch:     &branch,
		Prompt:     "fix it",
		Runtime:    &gqlclient.AgentRunMinimalFragment_Runtime{Type: provider},
		Upload:     &gqlclient.AgentRunMinimalFragment_Upload{Session: &session},
	}
	if prRef != "" {
		ref := prRef
		run.PullRequests = []*gqlclient.AgentRunMinimalFragment_PullRequests{{ID: "pr-1", Ref: &ref}}
	}
	return run
}

func TestResumeDownloadsAndRestoresSession(t *testing.T) {
	session := &fakeSession{}
	service := NewService(fakeResolver{})
	service.session = session
	service.newClient = func(string, string) (API, error) {
		return fakeAPI{runs: []*gqlclient.AgentRunMinimalFragment{resumableRun("run-1", "feat/fix")}}, nil
	}
	if err := service.Resume(t.Context(), "run-1", "/work/repo", "feat/fix"); err != nil {
		t.Fatal(err)
	}
	if session.runID != "run-1" || session.path != "/work/repo" {
		t.Fatalf("session = %#v", session)
	}
}

func TestResumeRequiresClonePath(t *testing.T) {
	service := NewService(fakeResolver{})
	err := service.Resume(t.Context(), "run-1", "  ", "")
	if !bridge.IsCode(err, bridge.ErrorInvalid) {
		t.Fatalf("err = %v", err)
	}
}

func TestResumeRejectsMissingSession(t *testing.T) {
	service := NewService(fakeResolver{})
	service.newClient = func(string, string) (API, error) {
		return fakeAPI{runs: []*gqlclient.AgentRunMinimalFragment{{ID: "missing"}}}, nil
	}
	err := service.Resume(t.Context(), "missing", "/work/repo", "")
	if !bridge.IsCode(err, bridge.ErrorUnavailable) {
		t.Fatalf("err = %v", err)
	}
}

func TestResumeReturnsSessionError(t *testing.T) {
	session := &fakeSession{err: errors.New("not a git checkout")}
	service := NewService(fakeResolver{})
	service.session = session
	service.newClient = func(string, string) (API, error) {
		return fakeAPI{runs: []*gqlclient.AgentRunMinimalFragment{resumableRun("run-1", "")}}, nil
	}
	if err := service.Resume(t.Context(), "run-1", "/work/repo", ""); err == nil || err.Error() != "not a git checkout" {
		t.Fatalf("err = %v", err)
	}
}

func TestApplyPullRequestKeepsMatchingRef(t *testing.T) {
	first, second := "feat/a", "feat/b"
	run := resumableRun("run-1", first)
	run.PullRequests = []*gqlclient.AgentRunMinimalFragment_PullRequests{
		{ID: "1", Ref: &first},
		{ID: "2", Ref: &second},
	}
	applyPullRequest(run, "feat/b")
	if len(run.PullRequests) != 1 || run.PullRequests[0].ID != "2" {
		t.Fatalf("prs = %#v", run.PullRequests)
	}
}
