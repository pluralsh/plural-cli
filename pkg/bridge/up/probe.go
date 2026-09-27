package up

import (
	"context"

	"github.com/pluralsh/plural-cli/pkg/provider"
)

// ProbeResult is the credential-checked provider setup payload for the Up wizard.
type ProbeResult struct {
	Summary string
	Fields  []FormField
}

// Prober checks cloud credentials and loads select options via provider.CloudSetup.
type Prober interface {
	Probe(ctx context.Context, providerID string) (ProbeResult, error)
	FieldOptions(ctx context.Context, providerID, fieldKey string, values map[string]string) ([]string, error)
	Preflights(ctx context.Context, providerID string, values map[string]string) error
}

// LiveProber delegates to each provider's CloudSetup implementation.
type LiveProber struct{}

// DefaultProber returns the live cloud prober.
func DefaultProber() Prober { return LiveProber{} }

// Probe validates credentials and returns form fields with select options filled.
func (LiveProber) Probe(ctx context.Context, providerID string) (ProbeResult, error) {
	setup, err := provider.Setup(providerID)
	if err != nil {
		return ProbeResult{}, err
	}
	res, err := setup.Probe(ctx)
	if err != nil {
		return ProbeResult{}, err
	}
	return ProbeResult{Summary: res.Summary, Fields: toFormFields(res.Fields)}, nil
}

// FieldOptions refreshes a dependent select via the provider CloudSetup.
func (LiveProber) FieldOptions(ctx context.Context, providerID, fieldKey string, values map[string]string) ([]string, error) {
	setup, err := provider.Setup(providerID)
	if err != nil {
		return nil, err
	}
	return setup.Options(ctx, fieldKey, values)
}

// Preflights runs provider.Preflights() for the surveyed values.
func (LiveProber) Preflights(ctx context.Context, providerID string, values map[string]string) error {
	setup, err := provider.Setup(providerID)
	if err != nil {
		return err
	}
	return setup.Preflights(ctx, values)
}

func toFormFields(in []provider.SetupField) []FormField {
	out := make([]FormField, len(in))
	for i, f := range in {
		out[i] = FormField{
			Key:         f.Key,
			Label:       f.Label,
			Placeholder: f.Placeholder,
			Default:     f.Default,
			Required:    f.Required,
			Options:     f.Options,
		}
	}
	return out
}
