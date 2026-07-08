package agents

import (
	"context"
	"testing"

	console "github.com/pluralsh/console/go/client"
)

type testRepository struct {
	prepared        bool
	validated       bool
	validatedBranch string
}

func (r *testRepository) Prepare(_ context.Context, _ *console.AgentRunMinimalFragment, _ *SessionBundle, _ string) (string, error) {
	r.prepared = true
	return "prepared-branch", nil
}

func (r *testRepository) ValidateRepository(_ string, _ *SessionManifest) error {
	return nil
}

func (r *testRepository) Validate(_ string, manifest *SessionManifest) error {
	r.validated = true
	r.validatedBranch = manifest.Branch
	return nil
}

type testRestorer struct{}

func (testRestorer) Provider() console.AgentRuntimeType {
	return console.AgentRuntimeTypeCodex
}

func (testRestorer) Prepare(_ context.Context, opts RestoreOptions) (*PreparedSession, error) {
	return &PreparedSession{
		RepoPath:  opts.RepoPath,
		WorkDir:   opts.WorkDir,
		SessionID: opts.Manifest.Session.ID,
	}, nil
}

func (testRestorer) Resume(_ context.Context, _ *PreparedSession) error {
	return nil
}

func TestSessionServiceRestoreAndResumeUsesRepositoryValidation(t *testing.T) {
	repository := &testRepository{}
	service := NewSessionService(WithSessionRepository(repository))
	service.registry = NewRestorerRegistry(testRestorer{})

	err := service.RestoreAndResume(context.Background(), &SessionBundle{
		Run: &console.AgentRunMinimalFragment{ID: "run-1"},
		Manifest: &SessionManifest{
			Provider:   console.AgentRuntimeTypeCodex,
			Repository: "git@github.com:pluralsh/plural.git",
			Session:    SessionMetadata{ID: "session-1"},
		},
		WorkDir: "/tmp",
	}, "/repo")
	if err != nil {
		t.Fatalf("RestoreAndResume returned error: %v", err)
	}
	if !repository.prepared {
		t.Fatalf("expected repository prepare to be called")
	}
	if !repository.validated {
		t.Fatalf("expected repository validation to be called")
	}
	if repository.validatedBranch != "prepared-branch" {
		t.Fatalf("expected validation to use prepared branch, got %q", repository.validatedBranch)
	}
}
