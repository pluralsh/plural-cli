package wkspace

import (
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
)

func (wk *Workspace) BuildManifest(prev *manifest.Manifest) *manifest.Manifest {
	repository := wk.Installation.Repository
	charts := make([]*manifest.ChartManifest, len(wk.Charts))
	terraform := make([]*manifest.TerraformManifest, len(wk.Terraform))

	for i, ci := range wk.Charts {
		charts[i] = buildChartManifest(ci)
	}
	for i, ti := range wk.Terraform {
		terraform[i] = buildTerraformManifest(ti)
	}

	return &manifest.Manifest{
		Id:           repository.Id,
		Name:         repository.Name,
		Cluster:      wk.Provider.Cluster(),
		Project:      wk.Provider.Project(),
		Bucket:       wk.Provider.Bucket(),
		Provider:     wk.Provider.Name(),
		Region:       wk.Provider.Region(),
		License:      wk.Installation.LicenseKey,
		Wait:         wk.requiresWait(),
		Charts:       charts,
		Terraform:    terraform,
		Dependencies: buildDependencies(repository.Name, wk.Charts, wk.Terraform),
		Context:      wk.Provider.Context(),
		Links:        prev.Links,
	}
}

func buildDependencies(repo string, charts []*api.ChartInstallation, tfs []*api.TerraformInstallation) []*manifest.Dependency {
	var deps []*manifest.Dependency
	var seen = make(map[string]bool)

	for _, chart := range charts {
		for _, dep := range chart.Chart.Dependencies.Dependencies {
			_, ok := seen[dep.Repo]
			if ok {
				continue
			}

			if dep.Repo != repo {
				deps = append(deps, &manifest.Dependency{Repo: dep.Repo})
				seen[dep.Repo] = true
			}
		}
	}

	for _, tf := range tfs {
		for _, dep := range tf.Terraform.Dependencies.Dependencies {
			_, ok := seen[dep.Repo]
			if ok {
				continue
			}

			if dep.Repo != repo {
				deps = append(deps, &manifest.Dependency{Repo: dep.Repo})
				seen[dep.Repo] = true
			}
		}
	}

	return deps
}

func buildChartManifest(chartInstallation *api.ChartInstallation) *manifest.ChartManifest {
	chart := chartInstallation.Chart
	version := chartInstallation.Version
	return &manifest.ChartManifest{Id: chart.Id, Name: chart.Name, VersionId: version.Id, Version: version.Version}
}

func buildTerraformManifest(tfInstallation *api.TerraformInstallation) *manifest.TerraformManifest {
	terraform := tfInstallation.Terraform
	return &manifest.TerraformManifest{Id: terraform.Id, Name: terraform.Name}
}
