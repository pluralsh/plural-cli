package api

import (
	"fmt"
	"os"
	"path"
)

type chartsResponse struct {
	Charts struct {
		Edges []*ChartEdge
	}
}

type versionsResponse struct {
	Versions struct {
		Edges []*VersionEdge
	}
}

type chartInstallationsResponse struct {
	ChartInstallations struct {
		Edges []*ChartInstallationEdge
	}
}

type packageCacheEntry struct {
	Charts    []*ChartInstallation
	Terraform []*TerraformInstallation
}

var packageCache = make(map[string]*packageCacheEntry)

var chartsQuery = fmt.Sprintf(`
	query ChartsQuery($id: ID!) {
		charts(repositoryId: $id, first: %d) {
			edges {
				node {
					...ChartFragment
				}
			}
		}
	}
	%s
`, pageSize, ChartFragment)

var versionsQuery = fmt.Sprintf(`
	query VersionsQuery($id: ID!) {
		versions(chartId: $id, first: %d) {
			edges {
				node {
					...VersionFragment
				}
			}
		}
	}
	%s
`, pageSize, VersionFragment)

const createCrdQuery = `
	mutation CrdCreate($chartName: ChartName!, $name: String!, $blob: UploadOrUrl!) {
		createCrd(chartName: $chartName, attributes: {name: $name, blob: $blob}) {
			id
		}
	}
`

var chartInstallationsQuery = fmt.Sprintf(`
	query CIQuery($id: ID!) {
		chartInstallations(repositoryId: $id, first: %d) {
			edges { node { ...ChartInstallationFragment } }
		}
	}
	%s
`, pageSize, ChartInstallationFragment)

var packageInstallationsQuery = fmt.Sprintf(`
	query Packages($id: ID!) {
		chartInstallations(repositoryId: $id, first: %d) {
			edges { node { ...ChartInstallationFragment } }
		}
		terraformInstallations(repositoryId: $id, first: %d) {
			edges { node { ...TerraformInstallationFragment } }
		}
	}
	%s
	%s
`, pageSize, pageSize, ChartInstallationFragment, TerraformInstallationFragment)

func (client *Client) GetCharts(repoId string) ([]*Chart, error) {
	var resp chartsResponse
	req := client.Build(chartsQuery)
	req.Var("id", repoId)
	err := client.Run(req, &resp)
	charts := make([]*Chart, len(resp.Charts.Edges))
	for i, edge := range resp.Charts.Edges {
		charts[i] = edge.Node
	}
	return charts, err
}

func (client *Client) GetVersions(chartId string) ([]*Version, error) {
	var resp versionsResponse
	req := client.Build(versionsQuery)
	req.Var("id", chartId)
	err := client.Run(req, &resp)
	versions := make([]*Version, len(resp.Versions.Edges))
	for i, edge := range resp.Versions.Edges {
		versions[i] = edge.Node
	}
	return versions, err
}

func (client *Client) GetChartInstallations(repoId string) ([]*ChartInstallation, error) {
	var resp chartInstallationsResponse
	req := client.Build(chartInstallationsQuery)
	req.Var("id", repoId)
	err := client.Run(req, &resp)
	insts := make([]*ChartInstallation, len(resp.ChartInstallations.Edges))
	for i, edge := range resp.ChartInstallations.Edges {
		insts[i] = edge.Node
	}
	return insts, err
}

func (client *Client) GetPackageInstallations(repoId string) (charts []*ChartInstallation, tfs []*TerraformInstallation, err error) {
	if entry, ok := packageCache[repoId]; ok {
		return entry.Charts, entry.Terraform, nil
	}

	var resp struct {
		ChartInstallations struct {
			Edges []*ChartInstallationEdge
		}
		TerraformInstallations struct {
			Edges []*TerraformInstallationEdge
		}
	}

	req := client.Build(packageInstallationsQuery)
	req.Var("id", repoId)
	err = client.Run(req, &resp)
	if err != nil {
		return
	}

	charts = make([]*ChartInstallation, len(resp.ChartInstallations.Edges))
	for i, edge := range resp.ChartInstallations.Edges {
		charts[i] = edge.Node
	}

	tfs = make([]*TerraformInstallation, len(resp.TerraformInstallations.Edges))
	for i, edge := range resp.TerraformInstallations.Edges {
		tfs[i] = edge.Node
	}

	if err == nil {
		packageCache[repoId] = &packageCacheEntry{Charts: charts, Terraform: tfs}
	}

	return
}

func (client *Client) CreateCrd(repo string, chart string, file string) error {
	var resp struct {
		Id string
	}
	name := path.Base(file)

	rf, err := os.Open(file)
	if err != nil {
		return err
	}
	defer rf.Close()

	req := client.Build(createCrdQuery)
	req.Var("chartName", ChartName{Chart: chart, Repo: repo})
	req.Var("name", name)
	req.Var("blob", "blob")
	req.File("blob", file, rf)

	return client.Run(req, &resp)
}
