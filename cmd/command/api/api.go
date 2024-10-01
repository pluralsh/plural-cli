package api

import (
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/polly/algorithms"
	"github.com/urfave/cli"
)

type Plural struct {
	client.Plural
}

func Command(clients client.Plural) cli.Command {
	plural := Plural{
		Plural: clients,
	}
	return cli.Command{
		Name:        "api",
		Usage:       "inspect the plural api",
		Subcommands: plural.apiCommands(),
		Category:    "API",
	}
}

func (p *Plural) apiCommands() []cli.Command {
	return []cli.Command{
		{
			Name:  "list",
			Usage: "lists plural resources",
			Subcommands: []cli.Command{
				{
					Name:   "installations",
					Usage:  "lists your installations",
					Action: common.LatestVersion(p.handleInstallations),
				},
				{
					Name:      "charts",
					Usage:     "lists charts for a repository",
					ArgsUsage: "{repository-id}",
					Action:    common.LatestVersion(common.RequireArgs(p.handleCharts, []string{"{repository-id}"})),
				},
				{
					Name:      "terraform",
					Usage:     "lists terraform modules for a repository",
					ArgsUsage: "{repository-id}",
					Action:    common.LatestVersion(common.RequireArgs(p.handleTerraforma, []string{"{repository-id}"})),
				},
				{
					Name:      "versions",
					Usage:     "lists versions of a chart",
					ArgsUsage: "{chart-id}",
					Action:    common.LatestVersion(common.RequireArgs(p.handleVersions, []string{"{repository-id}"})),
				},
				{
					Name:      "chartinstallations",
					Aliases:   []string{"ci"},
					Usage:     "lists chart installations for a repository",
					ArgsUsage: "{repository-id}",
					Action:    common.LatestVersion(common.RequireArgs(p.handleChartInstallations, []string{"{repository-id}"})),
				},
				{
					Name:      "terraforminstallations",
					Aliases:   []string{"ti"},
					Usage:     "lists terraform installations for a repository",
					ArgsUsage: "{repository-id}",
					Action:    common.LatestVersion(common.RequireArgs(p.handleTerraformInstallations, []string{"{repository-id}"})),
				},
				{
					Name:      "artifacts",
					Usage:     "Lists artifacts for a repository",
					ArgsUsage: "{repository-id}",
					Action:    common.LatestVersion(common.RequireArgs(p.handleArtifacts, []string{"{repository-id}"})),
				},
			},
		},
		{
			Name:  "create",
			Usage: "creates plural resources",
			Subcommands: []cli.Command{
				{
					Name:      "domain",
					Usage:     "creates a new domain for your account",
					ArgsUsage: "{domain}",
					Action:    common.LatestVersion(common.RequireArgs(p.handleCreateDomain, []string{"{domain}"})),
				},
			},
		},
	}
}

func (p *Plural) handleInstallations(c *cli.Context) error {
	p.InitPluralClient()
	installations, err := p.GetInstallations()
	if err != nil {
		return api.GetErrorResponse(err, "GetInstallations")
	}

	installations = algorithms.Filter(installations, func(v *api.Installation) bool {
		return v.Repository != nil
	})

	headers := []string{"Repository", "Repository Id", "Publisher"}
	return utils.PrintTable(installations, headers, func(inst *api.Installation) ([]string, error) {
		repo := inst.Repository
		publisherName := ""
		if repo.Publisher != nil {
			publisherName = repo.Publisher.Name
		}
		return []string{repo.Name, repo.Id, publisherName}, nil
	})
}

func (p *Plural) handleCharts(c *cli.Context) error {
	p.InitPluralClient()
	charts, err := p.GetCharts(c.Args().First())
	if err != nil {
		return api.GetErrorResponse(err, "GetCharts")
	}

	headers := []string{"Id", "Name", "Description", "Latest Version"}
	return utils.PrintTable(charts, headers, func(c *api.Chart) ([]string, error) {
		return []string{c.Id, c.Name, c.Description, c.LatestVersion}, nil
	})
}

func (p *Plural) handleTerraforma(c *cli.Context) error {
	p.InitPluralClient()
	tfs, err := p.GetTerraform(c.Args().First())
	if err != nil {
		return api.GetErrorResponse(err, "GetTerraforma")
	}

	headers := []string{"Id", "Name", "Description"}
	return utils.PrintTable(tfs, headers, func(tf *api.Terraform) ([]string, error) {
		return []string{tf.Id, tf.Name, tf.Description}, nil
	})
}

func (p *Plural) handleVersions(c *cli.Context) error {
	p.InitPluralClient()
	versions, err := p.GetVersions(c.Args().First())
	if err != nil {
		return api.GetErrorResponse(err, "GetVersions")
	}

	headers := []string{"Id", "Version"}
	return utils.PrintTable(versions, headers, func(v *api.Version) ([]string, error) {
		return []string{v.Id, v.Version}, nil
	})
}

func (p *Plural) handleChartInstallations(c *cli.Context) error {
	p.InitPluralClient()
	chartInstallations, err := p.GetChartInstallations(c.Args().First())
	if err != nil {
		return api.GetErrorResponse(err, "GetChartInstallations")
	}

	cis := algorithms.Filter(chartInstallations, func(ci *api.ChartInstallation) bool {
		return ci.Chart != nil && ci.Version != nil
	})

	row := func(ci *api.ChartInstallation) ([]string, error) {
		return []string{ci.Id, ci.Chart.Id, ci.Chart.Name, ci.Version.Version}, nil
	}
	headers := []string{"Id", "Chart Id", "Chart Name", "Version"}
	return utils.PrintTable(cis, headers, row)
}

func (p *Plural) handleTerraformInstallations(c *cli.Context) error {
	p.InitPluralClient()
	terraformInstallations, err := p.GetTerraformInstallations(c.Args().First())
	if err != nil {
		return api.GetErrorResponse(err, "GetTerraformInstallations")
	}

	tis := algorithms.Filter(terraformInstallations, func(ti *api.TerraformInstallation) bool {
		return ti != nil
	})

	headers := []string{"Id", "Terraform Id", "Name"}
	return utils.PrintTable(tis, headers, func(ti *api.TerraformInstallation) ([]string, error) {
		tf := ti.Terraform
		return []string{ti.Id, tf.Id, tf.Name}, nil
	})
}

func (p *Plural) handleArtifacts(c *cli.Context) error {
	p.InitPluralClient()
	artifacts, err := p.ListArtifacts(c.Args().First())
	if err != nil {
		return api.GetErrorResponse(err, "ListArtifacts")
	}

	headers := []string{"Id", "Name", "Platform", "Blob", "Sha"}
	return utils.PrintTable(artifacts, headers, func(art api.Artifact) ([]string, error) {
		return []string{art.Id, art.Name, art.Platform, art.Blob, art.Sha}, nil
	})
}

func (p *Plural) handleCreateDomain(c *cli.Context) error {
	p.InitPluralClient()
	err := p.CreateDomain(c.Args().First())
	return api.GetErrorResponse(err, "CreateDomain")
}
