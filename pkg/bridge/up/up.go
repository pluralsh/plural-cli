// Package up exposes credential-free setup helpers for the Up wizard.
package up

import (
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/provider"
)

// Flow is one top-level plural-up path (maps to --cloud / --dry-run).
type Flow struct {
	ID     string
	Title  string
	Blurb  string
	Cloud  bool
	DryRun bool
}

// NeedsProvider is true when this flow runs the provider survey (CLI GetProvider
// via HandleInitWithProject). Cloud runs that survey after Console instance pick;
// dry-run-only stubs still skip it in the TUI for now.
func (f Flow) NeedsProvider() bool {
	return f.ID == "self-hosted" || f.ID == "cloud"
}

// CLI returns the equivalent plural up invocation for this flow.
func (f Flow) CLI(ignorePreflights bool) string {
	cmd := "plural up"
	if f.Cloud {
		cmd += " --cloud"
	}
	if f.DryRun {
		cmd += " --dry-run"
	}
	if ignorePreflights {
		cmd += " --ignore-preflights"
	}
	return cmd
}

// Flows returns the setup modes offered on the first Up screen.
func Flows() []Flow {
	return []Flow{
		{
			ID:    "self-hosted",
			Title: "Self-hosted",
			Blurb: "pick a cloud provider · provision management cluster",
		},
		{
			ID:    "cloud",
			Title: "Plural Cloud",
			Blurb: "pick a Console instance (--cloud) · then provider survey",
			Cloud: true,
		},
		{
			ID:     "dry-run",
			Title:  "Dry-run",
			Blurb:  "generate repo only (--dry-run) · coming next",
			DryRun: true,
		},
		{
			ID:     "cloud-dry-run",
			Title:  "Cloud · dry-run",
			Blurb:  "Plural Cloud generate only · coming next",
			Cloud:  true,
			DryRun: true,
		},
	}
}

// Provider is one cloud target selectable for self-hosted up.
type Provider struct {
	ID    string
	Title string
	Blurb string
}

// CloudProviders returns the providers offered by self-hosted `plural up` init.
func CloudProviders() []Provider {
	out := make([]Provider, 0, 4)
	for _, s := range provider.Setups() {
		switch s.Name() {
		case api.ProviderAWS:
			out = append(out, Provider{ID: s.Name(), Title: "AWS", Blurb: "Amazon Web Services"})
		case api.ProviderAzure:
			out = append(out, Provider{ID: s.Name(), Title: "Azure", Blurb: "Microsoft Azure"})
		case api.ProviderGCP:
			out = append(out, Provider{ID: s.Name(), Title: "GCP", Blurb: "Google Cloud Platform"})
		case api.BYOK:
			out = append(out, Provider{ID: s.Name(), Title: "BYOK", Blurb: "bring your own Kubernetes cluster"})
		}
	}
	return out
}

// FormField describes one provider-setup input (CLI survey parity).
// When Options is non-empty the TUI renders a select (survey.Select parity).
type FormField struct {
	Key         string
	Label       string
	Placeholder string
	Default     string
	Required    bool
	Options     []string
}

// ProviderFormFields returns the self-hosted init fields for a provider.
// Sourced from each provider's CloudSetup schema.
func ProviderFormFields(providerID string) []FormField {
	setup, err := provider.Setup(providerID)
	if err != nil {
		return nil
	}
	return toFormFields(setup.Schema())
}
