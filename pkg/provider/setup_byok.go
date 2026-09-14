package provider

import "context"

type byokSetup struct{}

func (byokSetup) Name() string { return "byok" }

func (byokSetup) Schema() []SetupField {
	return []SetupField{
		{Key: "cluster", Label: "Cluster name", Placeholder: "name for this cluster", Required: true},
		{Key: "kubeconfig", Label: "Kubeconfig path", Placeholder: "~/.kube/config", Default: "~/.kube/config", Required: true},
		{Key: "database", Label: "Console DB URL", Placeholder: "postgres://user:pass@host:5432/db", Required: true},
		{Key: "domain", Label: "Console domain", Placeholder: "console.example.com", Required: true},
	}
}

func (s byokSetup) Probe(context.Context) (SetupResult, error) {
	return SetupResult{
		Summary: "BYOK uses your local kubeconfig (checked on deploy).",
		Fields:  s.Schema(),
	}, nil
}

func (byokSetup) Options(context.Context, string, map[string]string) ([]string, error) {
	return nil, nil
}

// Preflights: cluster connectivity needs a configured ByokProvider; deferred to deploy.
func (byokSetup) Preflights(context.Context, map[string]string) error { return nil }
