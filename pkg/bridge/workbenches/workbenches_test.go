package workbenches

import (
	"context"
	"testing"
	"time"

	gqlclient "github.com/pluralsh/console/go/client"

	"github.com/pluralsh/plural-cli/pkg/console"
)

type fakeResolver struct{}

func (fakeResolver) ActiveConsole(context.Context) (string, string, error) {
	return "https://console.example.com", "token", nil
}

type fakeAPI struct{}

func (fakeAPI) ListWorkbenches(*string, *int64, *string) (*gqlclient.ListWorkbenches_Workbenches, error) {
	return nil, nil
}
func (fakeAPI) ListWorkbenchJobs(string, int, int) ([]console.WorkbenchJob, error) { return nil, nil }
func (fakeAPI) CreateQueuedPrompt(string, string, time.Time) (*gqlclient.QueuedPromptFragment, error) {
	return &gqlclient.QueuedPromptFragment{ID: "prompt-1"}, nil
}

func TestFollowUpQueuesPromptForSelectedJob(t *testing.T) {
	service := NewService(fakeResolver{})
	service.newClient = func(string, string) (API, error) { return fakeAPI{}, nil }
	result, err := service.FollowUp(t.Context(), "job-1", "verify the fix", 0)
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != "prompt-1" || result.WorkbenchID != "job-1" {
		t.Fatalf("unexpected result: %#v", result)
	}
}
