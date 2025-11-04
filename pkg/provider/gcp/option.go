package gcp

import (
	"fmt"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/manifest"
)

type Option func(*Provider) error

func WithConfig(c config.Config, defaultCluster string, cloudEnabled bool) Option {
	return func(gcp *Provider) error {
		inputProvider, err := NewSurvey(defaultCluster)
		if err != nil {
			return err
		}

		gcp.InputProvider = inputProvider
		gcp.ctx = map[string]interface{}{
			"BucketLocation": getBucketLocation(gcp.Region()),
			// Location might conflict with the region set by users. However, this is only a temporary solution that should be removed
			"Location": gcp.Region(),
		}

		projectManifest := manifest.ProjectManifest{
			Cluster:  gcp.Cluster(),
			Project:  gcp.Project(),
			Provider: api.ProviderGCP,
			Region:   gcp.Region(),
			Context:  gcp.Context(),
			Owner:    &manifest.Owner{Email: c.Email, Endpoint: c.Endpoint},
		}

		gcp.writer = projectManifest.Configure(cloudEnabled, gcp.Cluster())
		gcp.bucket = projectManifest.Bucket

		return nil
	}
}

func WithManifest(m *manifest.ProjectManifest) Option {
	return func(gcp *Provider) error {
		// Needed to update legacy deployments
		if m.Region == "" {
			m.Region = "us-east1"
			if err := m.Write(manifest.ProjectManifestPath()); err != nil {
				return err
			}
		} else if location := strings.Split(m.Region, "-"); len(location) >= 3 {
			m.Context["Location"] = m.Region
			m.Region = fmt.Sprintf("%s-%s", location[0], location[1])
			m.Context["BucketLocation"] = getBucketLocation(m.Region)
			if err := m.Write(manifest.ProjectManifestPath()); err != nil {
				return err
			}
		}
		// Needed to update legacy deployments
		if _, ok := m.Context["BucketLocation"]; !ok {
			m.Context["BucketLocation"] = "US"
			if err := m.Write(manifest.ProjectManifestPath()); err != nil {
				return err
			}
		}
		// Needed to update legacy deployments
		if _, ok := m.Context["Location"]; !ok {
			m.Context["Location"] = m.Region
			if err := m.Write(manifest.ProjectManifestPath()); err != nil {
				return err
			}
		}

		gcp.bucket = m.Bucket
		gcp.ctx = m.Context
		gcp.InputProvider = NewReadonlyInputProvider(m.Cluster, m.Project, m.Region)

		return nil
	}
}
