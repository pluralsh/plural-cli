package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/provider/gcp"
)

type gcpSetup struct{}

func (gcpSetup) Name() string { return "gcp" }

func (gcpSetup) Schema() []SetupField {
	return []SetupField{
		{Key: "cluster", Label: "Cluster name", Placeholder: "max 15 chars", Required: true},
		{Key: "project", Label: "GCP project ID", Placeholder: "your GCP project", Required: true},
		{Key: "region", Label: "Region", Placeholder: "e.g. us-east1", Default: "us-east1", Required: true},
	}
}

func (s gcpSetup) Probe(ctx context.Context) (SetupResult, error) {
	email, name, err := gcp.LoggedInUserInfo()
	if err != nil {
		return SetupResult{}, fmt.Errorf("GCP credentials: %w", err)
	}
	projects, err := gcp.Projects()
	if err != nil {
		return SetupResult{}, fmt.Errorf("GCP projects: %w", err)
	}

	fields := s.Schema()
	fields = withOptions(fields, "project", projects)
	for i := range fields {
		if fields[i].Key == "region" {
			// Same as CLI survey: regions load after project (gcp.Regions).
			fields[i].Options = nil
			fields[i].Placeholder = "select a project first"
		}
	}

	summary := email
	if name != "" {
		summary = fmt.Sprintf("%s (%s)", email, name)
	}
	return SetupResult{Summary: "GCP · " + summary, Fields: fields}, nil
}

func (gcpSetup) Options(_ context.Context, fieldKey string, values map[string]string) ([]string, error) {
	switch fieldKey {
	case "project":
		return gcp.Projects()
	case "region":
		return gcp.Regions(strings.TrimSpace(values["project"])), nil
	default:
		return nil, nil
	}
}

// Preflights runs enabled-services + permissions checks like gcp.Provider.Preflights.
func (gcpSetup) Preflights(ctx context.Context, values map[string]string) error {
	project := strings.TrimSpace(values["project"])
	region := strings.TrimSpace(values["region"])
	cluster := strings.TrimSpace(values["cluster"])
	if project == "" {
		return fmt.Errorf("GCP project is required for preflights")
	}
	prov, err := gcp.NewProvider(gcp.WithManifest(&manifest.ProjectManifest{
		Cluster:  cluster,
		Project:  project,
		Provider: api.ProviderGCP,
		Region:   region,
		Context:  map[string]interface{}{"Location": region, "BucketLocation": "US"},
	}))
	if err != nil {
		return err
	}
	for _, pre := range prov.Preflights() {
		if err := pre.Validate(); err != nil {
			return err
		}
	}
	return nil
}
