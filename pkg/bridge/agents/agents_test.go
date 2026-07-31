package agents

import (
	"context"
	"testing"

	gqlclient "github.com/pluralsh/console/go/client"
)

type fakeResolver struct{}

func (fakeResolver) ActiveConsole(context.Context) (string, string, error) {
	return "https://console.example.com", "token", nil
}

type fakeAPI struct {
	runs []*gqlclient.AgentRunMinimalFragment
}

func (f fakeAPI) ListAgentRuns(int64) ([]*gqlclient.AgentRunMinimalFragment, error) {
	return f.runs, nil
}
func (f fakeAPI) GetAgentRun(id string) (*gqlclient.AgentRunMinimalFragment, error) {
	for _, run := range f.runs {
		if run.ID == id {
			return run, nil
		}
	}
	return nil, nil
}

func TestListOnlyReturnsResumableRuns(t *testing.T) {
	session := "https://example.com/session.tgz"
	provider := gqlclient.AgentRuntimeTypeCodex
	branch := "main"
	service := NewService(fakeResolver{})
	service.newClient = func(string, string) (API, error) {
		return fakeAPI{runs: []*gqlclient.AgentRunMinimalFragment{
			{ID: "ready", Repository: "git@github.com:acme/repo.git", Branch: &branch, Prompt: "fix it", Runtime: &gqlclient.AgentRunMinimalFragment_Runtime{Type: provider}, Upload: &gqlclient.AgentRunMinimalFragment_Upload{Session: &session}},
			{ID: "missing", Repository: "repo"},
		}}, nil
	}
	page, err := service.List(t.Context(), nil, "acme")
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != "ready" {
		t.Fatalf("unexpected page: %#v", page)
	}
}
