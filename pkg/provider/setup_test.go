package provider

import (
	"context"
	"testing"
)

func TestSetupRegistry(t *testing.T) {
	want := []string{"aws", "azure", "gcp", "byok"}
	got := Setups()
	if len(got) != len(want) {
		t.Fatalf("len = %d", len(got))
	}
	for i, id := range want {
		if got[i].Name() != id {
			t.Fatalf("Setups()[%d] = %q", i, got[i].Name())
		}
		s, err := Setup(id)
		if err != nil {
			t.Fatalf("Setup(%q): %v", id, err)
		}
		if len(s.Schema()) == 0 {
			t.Fatalf("%s schema empty", id)
		}
		opts, err := s.Options(context.Background(), "missing", nil)
		if err != nil {
			t.Fatalf("%s Options: %v", id, err)
		}
		_ = opts
	}
	if _, err := Setup("nope"); err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestAWSSchemaHasRegion(t *testing.T) {
	s, err := Setup("aws")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range s.Schema() {
		if f.Key == "region" {
			found = true
		}
	}
	if !found {
		t.Fatal("aws schema missing region")
	}
	opts, err := s.Options(context.Background(), "region", nil)
	if err != nil || len(opts) < 5 {
		t.Fatalf("aws regions = %#v err=%v", opts, err)
	}
}
