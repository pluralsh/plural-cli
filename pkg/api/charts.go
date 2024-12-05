package api

import (
	"github.com/pluralsh/gqlclient"
	"github.com/pluralsh/gqlclient/pkg/utils"
)

func convertVersion(version *gqlclient.VersionFragment) *Version {
	if version == nil {
		return nil
	}
	v := &Version{
		Id:      version.ID,
		Version: version.Version,
		Helm:    version.Helm,
	}
	if version.Readme != nil {
		v.Readme = *version.Readme
	}
	if version.Package != nil {
		v.Package = *version.Package
	}
	if version.ValuesTemplate != nil {
		v.ValuesTemplate = *version.ValuesTemplate
	}
	v.TemplateType = gqlclient.TemplateTypeGotemplate
	if version.TemplateType != nil {
		v.TemplateType = *version.TemplateType
	}
	if version.InsertedAt != nil {
		v.InsertedAt = *version.InsertedAt
	}

	v.Crds = make([]Crd, 0)
	for _, crd := range version.Crds {
		v.Crds = append(v.Crds, convertCrd(crd))
	}
	v.Dependencies = convertDependencies(version.Dependencies)
	return v
}

func convertCrd(crd *gqlclient.CrdFragment) Crd {
	c := Crd{
		Id:   crd.ID,
		Name: crd.Name,
		Blob: utils.ConvertStringPointer(crd.Blob),
	}

	return c
}

func convertDependencies(depFragment *gqlclient.DependenciesFragment) *Dependencies {
	if depFragment == nil {
		return nil
	}
	dep := &Dependencies{
		Outputs:         depFragment.Outputs,
		Secrets:         utils.ConvertStringArrayPointer(depFragment.Secrets),
		Providers:       convertProviders(depFragment.Providers),
		ProviderWirings: depFragment.ProviderWirings,
	}
	if depFragment.ProviderVsn != nil {
		dep.ProviderVsn = *depFragment.ProviderVsn
	}
	if depFragment.CliVsn != nil {
		dep.CliVsn = *depFragment.CliVsn
	}
	if depFragment.Application != nil {
		dep.Application = *depFragment.Application
	}
	if depFragment.Wait != nil {
		dep.Wait = *depFragment.Wait
	}
	dep.Dependencies = make([]*Dependency, 0)
	for _, dependency := range depFragment.Dependencies {
		dep.Dependencies = append(dep.Dependencies, &Dependency{
			Type: string(*dependency.Type),
			Repo: utils.ConvertStringPointer(dependency.Repo),
			Name: utils.ConvertStringPointer(dependency.Name),
		})
	}
	if depFragment.Wirings != nil {
		dep.Wirings = &Wirings{
			Terraform: utils.ConvertMapInterfaceToString(depFragment.Wirings.Terraform),
			Helm:      utils.ConvertMapInterfaceToString(depFragment.Wirings.Helm),
		}
	}

	return dep
}

func convertProviders(providers []*gqlclient.Provider) []string {
	p := make([]string, 0)
	for _, provider := range providers {
		p = append(p, string(*provider))
	}

	return p
}

func convertChartInstallation(fragment *gqlclient.ChartInstallationFragment) *ChartInstallation {
	if fragment == nil {
		return nil
	}
	return &ChartInstallation{
		Id: *fragment.ID,
		Chart: &Chart{
			Id:            utils.ConvertStringPointer(fragment.Chart.ID),
			Name:          fragment.Chart.Name,
			Description:   utils.ConvertStringPointer(fragment.Chart.Description),
			LatestVersion: utils.ConvertStringPointer(fragment.Chart.LatestVersion),
			Dependencies:  convertDependencies(fragment.Chart.Dependencies),
		},
		Version: convertVersion(fragment.Version),
	}
}
