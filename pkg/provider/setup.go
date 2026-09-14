package provider

import (
	"context"
	"fmt"

	"github.com/pluralsh/plural-cli/pkg/api"
)

// SetupField is one plural-up survey field (input or select).
// Options non-empty means survey.Select parity.
type SetupField struct {
	Key         string
	Label       string
	Placeholder string
	Default     string
	Required    bool
	Options     []string
}

// SetupResult is the credential-checked setup payload for a cloud provider.
type SetupResult struct {
	Summary string
	Fields  []SetupField
}

// CloudSetup is implemented by each cloud provider for the non-interactive
// half of plural up init: verify credentials and load select options
// (regions, projects, …) the same way mkAWS / mkAzure / GCP survey do.
//
// CLI order (common.RunPreflights): Probe/survey first, then Preflights().
// --ignore-preflights only skips Preflights() failures after a successful survey.
type CloudSetup interface {
	Name() string
	Schema() []SetupField
	Probe(ctx context.Context) (SetupResult, error)
	Options(ctx context.Context, fieldKey string, values map[string]string) ([]string, error)
	// Preflights runs provider.Preflights() checks after the survey fields are known.
	Preflights(ctx context.Context, values map[string]string) error
}

// Setup returns the CloudSetup implementation for a provider id (aws/azure/gcp/byok).
func Setup(name string) (CloudSetup, error) {
	switch name {
	case api.ProviderAWS:
		return awsSetup{}, nil
	case api.ProviderAzure:
		return azureSetup{}, nil
	case api.ProviderGCP:
		return gcpSetup{}, nil
	case api.BYOK:
		return byokSetup{}, nil
	default:
		return nil, fmt.Errorf("unknown provider %q", name)
	}
}

// Setups returns CloudSetup for every self-hosted up provider.
func Setups() []CloudSetup {
	return []CloudSetup{awsSetup{}, azureSetup{}, gcpSetup{}, byokSetup{}}
}

func withOptions(fields []SetupField, key string, options []string) []SetupField {
	out := make([]SetupField, len(fields))
	copy(out, fields)
	for i := range out {
		if out[i].Key == key {
			out[i].Options = options
		}
	}
	return out
}

func truncateMiddle(v string, n int) string {
	if n < 8 || len(v) <= n {
		return v
	}
	keep := (n - 1) / 2
	return v[:keep] + "…" + v[len(v)-keep:]
}
