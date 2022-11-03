package main

import (
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/containers"
	"github.com/urfave/cli"
)

func (p *Plural) apiCommands() []cli.Command {
	return []cli.Command{
		{
			Name:  "list",
			Usage: "lists plural resources",
			Subcommands: []cli.Command{
				{
					Name:      "installations",
					Usage:     "lists your installations",
					ArgsUsage: "",
					Action:    p.handleInstallations,
				},
				{
					Name:      "charts",
					Usage:     "lists charts for a repository",
					ArgsUsage: "REPO_ID",
					Action:    requireArgs(p.handleCharts, []string{"REPO_ID"}),
				},
				{
					Name:      "terraform",
					Usage:     "lists terraform modules for a repository",
					ArgsUsage: "REPO_ID",
					Action:    requireArgs(p.handleTerraforma, []string{"REPO_ID"}),
				},
				{
					Name:      "versions",
					Usage:     "lists versions of a chart",
					ArgsUsage: "CHART_ID",
					Action:    requireArgs(p.handleVersions, []string{"CHART_ID"}),
				},
				{
					Name:      "chartinstallations",
					Aliases:   []string{"ci"},
					Usage:     "lists chart installations for a repository",
					ArgsUsage: "REPO_ID",
					Action:    requireArgs(p.handleChartInstallations, []string{"REPO_ID"}),
				},
				{
					Name:      "terraforminstallations",
					Aliases:   []string{"ti"},
					Usage:     "lists terraform installations for a repository",
					ArgsUsage: "REPO_ID",
					Action:    requireArgs(p.handleTerraformInstallations, []string{"REPO_ID"}),
				},
				{
					Name:      "artifacts",
					Usage:     "Lists artifacts for a repository",
					ArgsUsage: "REPO_ID",
					Action:    requireArgs(p.handleArtifacts, []string{"REPO_ID"}),
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
					ArgsUsage: "DOMAIN",
					Action:    p.handleCreateDomain,
				},
			},
		},
	}
}

func (p *Plural) handleInstallations(c *cli.Context) error {
	p.InitPluralClient()
	installations, err := p.GetInstallations()
	if err != nil {
		return err
	}

	installations = containers.Filter(installations, func(v *api.Installation) bool {
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
		return err
	}

	headers := []string{"Id", "Name", "Description", "Latest Version"}
	return utils.PrintTable(charts, headers, func(c *api.Chart) ([]string, error) {
		return []string{c.Id, c.Name, c.Description, c.LatestVersion}, nil
	})
}

func (p *Plural) handleTerraforma(c *cli.Context) error {
	p.InitPluralClient()
	tfs, err := p.GetTerraforma(c.Args().First())
	if err != nil {
		return err
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
		return err
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
		return err
	}

	cis := containers.Filter(chartInstallations, func(ci *api.ChartInstallation) bool {
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
		return err
	}

	tis := containers.Filter(terraformInstallations, func(ti *api.TerraformInstallation) bool {
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
		return err
	}

	headers := []string{"Id", "Name", "Platform", "Blob", "Sha"}
	return utils.PrintTable(artifacts, headers, func(art api.Artifact) ([]string, error) {
		return []string{art.Id, art.Name, art.Platform, art.Blob, art.Sha}, nil
	})
}

func (p *Plural) handleCreateDomain(c *cli.Context) error {
	p.InitPluralClient()
	return p.CreateDomain(c.Args().First())
}
