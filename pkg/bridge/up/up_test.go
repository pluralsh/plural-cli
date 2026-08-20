package up

import "testing"

func TestFlows(t *testing.T) {
	flows := Flows()
	if len(flows) != 4 {
		t.Fatalf("len = %d", len(flows))
	}
	if !flows[0].NeedsProvider() {
		t.Fatal("self-hosted should need provider")
	}
	if !flows[1].NeedsProvider() {
		t.Fatal("cloud should need provider after Console pick")
	}
	for _, f := range flows[2:] {
		if f.NeedsProvider() {
			t.Fatalf("%s should not need provider list yet", f.ID)
		}
	}
	want := []struct {
		id, cli       string
		cloud, dryRun bool
	}{
		{"self-hosted", "plural up", false, false},
		{"cloud", "plural up --cloud", true, false},
		{"dry-run", "plural up --dry-run", false, true},
		{"cloud-dry-run", "plural up --cloud --dry-run", true, true},
	}
	for i, w := range want {
		f := flows[i]
		if f.ID != w.id || f.Cloud != w.cloud || f.DryRun != w.dryRun || f.CLI(false) != w.cli {
			t.Fatalf("flows[%d] = %#v cli=%q", i, f, f.CLI(false))
		}
	}
	if got := flows[0].CLI(true); got != "plural up --ignore-preflights" {
		t.Fatalf("ignore-preflights cli = %q", got)
	}
	if got := flows[1].CLI(true); got != "plural up --cloud --ignore-preflights" {
		t.Fatalf("cloud ignore cli = %q", got)
	}
}

func TestCloudProviders(t *testing.T) {
	providers := CloudProviders()
	if len(providers) != 4 {
		t.Fatalf("len = %d", len(providers))
	}
	want := []string{"aws", "azure", "gcp", "byok"}
	for i, id := range want {
		if providers[i].ID != id || providers[i].Title == "" {
			t.Fatalf("providers[%d] = %#v", i, providers[i])
		}
	}
}

func TestProviderFormFields(t *testing.T) {
	if got := ProviderFormFields("aws"); len(got) != 2 || got[0].Key != "cluster" || got[1].Key != "region" {
		t.Fatalf("aws fields = %#v", got)
	}
	if got := ProviderFormFields("byok"); len(got) != 4 {
		t.Fatalf("byok fields = %#v", got)
	}
}

func TestValidateProviderForm(t *testing.T) {
	if err := ValidateProviderForm("aws", map[string]string{"cluster": "demo", "region": "us-east-2"}); err != nil {
		t.Fatalf("valid aws: %v", err)
	}
	if err := ValidateProviderForm("aws", map[string]string{"cluster": "this-name-is-way-too-long", "region": "us-east-2"}); err == nil {
		t.Fatal("expected cluster length error")
	}
	if err := ValidateProviderForm("aws", map[string]string{"cluster": "demo", "region": ""}); err == nil {
		t.Fatal("expected region required")
	}
}
