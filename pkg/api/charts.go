package api

import (
	"context"
	"os"
	"path"

	"github.com/pluralsh/gqlclient"
	"github.com/pluralsh/gqlclient/pkg/utils"
)

type packageCacheEntry struct {
	Charts    []*ChartInstallation
	Terraform []*TerraformInstallation
}

var packageCache = make(map[string]*packageCacheEntry)

func (client *client) GetCharts(repoId string) ([]*Chart, error) {
	charts := make([]*Chart, 0)
	resp, err := client.pluralClient.GetCharts(client.ctx, repoId)
	if err != nil {
		return nil, err
	}
	for _, edge := range resp.Charts.Edges {
		charts = append(charts, &Chart{
			Id:            utils.ConvertStringPointer(edge.Node.ID),
			Name:          edge.Node.Name,
			Description:   utils.ConvertStringPointer(edge.Node.Description),
			LatestVersion: utils.ConvertStringPointer(edge.Node.LatestVersion),
		})
	}

	return charts, err
}

func (client *client) GetVersions(chartId string) ([]*Version, error) {
	versions := make([]*Version, 0)
	resp, err := client.pluralClient.GetVersions(client.ctx, chartId)
	if err != nil {
		return nil, err
	}
	for _, version := range resp.Versions.Edges {
		versions = append(versions, convertVersion(version.Node))
	}
	return versions, err
}

func (client *client) GetChartInstallations(repoId string) ([]*ChartInstallation, error) {
	insts := make([]*ChartInstallation, 0)
	resp, err := client.pluralClient.GetChartInstallations(client.ctx, repoId)
	if err != nil {
		return nil, err
	}

	for _, edge := range resp.ChartInstallations.Edges {
		if edge.Node != nil {
			insts = append(insts, convertChartInstallation(edge.Node))
		}
	}

	return insts, err
}

func (client *client) GetPackageInstallations(repoId string) (charts []*ChartInstallation, tfs []*TerraformInstallation, err error) {
	if entry, ok := packageCache[repoId]; ok {
		return entry.Charts, entry.Terraform, nil
	}

	resp, err := client.pluralClient.GetPackageInstallations(client.ctx, repoId)
	if err != nil {
		return
	}

	charts = make([]*ChartInstallation, 0)
	for _, edge := range resp.ChartInstallations.Edges {
		if edge.Node != nil {
			charts = append(charts, convertChartInstallation(edge.Node))
		}
	}

	tfs = make([]*TerraformInstallation, 0)
	for _, edge := range resp.TerraformInstallations.Edges {
		node := edge.Node
		if node != nil {
			tfInstall := &TerraformInstallation{
				Id:        utils.ConvertStringPointer(node.ID),
				Terraform: convertTerraform(node.Terraform),

				Version: convertVersion(node.Version),
			}

			tfs = append(tfs, tfInstall)
		}
	}

	if err == nil {
		packageCache[repoId] = &packageCacheEntry{Charts: charts, Terraform: tfs}
	}

	return
}

func (client *client) CreateCrd(repo string, chart string, file string) error {
	name := path.Base(file)

	rf, err := os.Open(file)
	if err != nil {
		return err
	}
	defer func(rf *os.File) {
		_ = rf.Close()
	}(rf)

	upload := gqlclient.Upload{
		R:     rf,
		Name:  file,
		Field: "blob",
	}

	_, err = client.pluralClient.CreateCrd(context.Background(), gqlclient.ChartName{
		Chart: &chart,
		Repo:  &repo,
	}, name, "blob", gqlclient.WithFiles([]gqlclient.Upload{upload}))

	return err
}

func (client *client) UninstallChart(id string) (err error) {
	_, err = client.pluralClient.UninstallChart(client.ctx, id)
	return
}

func convertVersion(version *gqlclient.VersionFragment) *Version {
	if version == nil {
		return nil
	}
	v := &Version{
		Id:      version.ID,
		Version: version.Version,
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
