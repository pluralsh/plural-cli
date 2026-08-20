package up

import (
	"context"
	"errors"
	"testing"

	"github.com/pluralsh/plural-cli/pkg/console"
)

func TestDefaultInstanceIndex(t *testing.T) {
	instances := []ConsoleInstance{
		{Name: "a", URL: "https://a.onplural.sh"},
		{Name: "b", URL: "https://b.onplural.sh"},
	}
	if got := DefaultInstanceIndex(instances, ""); got != 0 {
		t.Fatalf("empty prior = %d", got)
	}
	if got := DefaultInstanceIndex(instances, "https://b.onplural.sh"); got != 1 {
		t.Fatalf("match b = %d", got)
	}
	if got := DefaultInstanceIndex(instances, "https://other.example"); got != 0 {
		t.Fatalf("no match = %d", got)
	}
}

func TestPriorConsoleMatches(t *testing.T) {
	if !PriorConsoleMatches("https://demo.onplural.sh", "https://demo.onplural.sh/gql") {
		t.Fatal("expected hostname match")
	}
	if PriorConsoleMatches("", "https://demo.onplural.sh") {
		t.Fatal("empty prior should not match")
	}
}

func TestValidateConsoleConfig(t *testing.T) {
	instances := []ConsoleInstance{{ID: "1", Name: "demo", URL: "https://demo.onplural.sh"}}
	if err := ValidateConsoleConfig(instances, console.Config{}); err == nil {
		t.Fatal("expected empty url error")
	}
	if err := ValidateConsoleConfig(instances, console.Config{Url: "https://demo.onplural.sh"}); err != nil {
		t.Fatalf("matching config: %v", err)
	}
	if err := ValidateConsoleConfig(instances, console.Config{Url: "https://other.onplural.sh"}); err == nil {
		t.Fatal("expected mismatch error")
	}
}

type stubLister struct {
	items []ConsoleInstance
	err   error
}

func (s stubLister) List(context.Context) ([]ConsoleInstance, error) {
	return s.items, s.err
}

func TestStubInstanceLister(t *testing.T) {
	l := stubLister{err: errors.New("boom")}
	if _, err := l.List(context.Background()); err == nil {
		t.Fatal("expected error")
	}
	l = stubLister{items: []ConsoleInstance{{Name: "x", URL: "https://x.onplural.sh"}}}
	got, err := l.List(context.Background())
	if err != nil || len(got) != 1 || got[0].Name != "x" {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestNeedsProviderIncludesCloud(t *testing.T) {
	for _, f := range Flows() {
		want := f.ID == "self-hosted" || f.ID == "cloud"
		if f.NeedsProvider() != want {
			t.Fatalf("%s NeedsProvider=%v want %v", f.ID, f.NeedsProvider(), want)
		}
	}
}
