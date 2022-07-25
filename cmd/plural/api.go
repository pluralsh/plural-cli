package main

import (
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

func (p *Plural) apiCommands() []cli.Command {
	return []cli.Command{
		{
			Name:  "list",
			Usage: "lists forge resources",
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
					Action:    p.handleCharts,
				},
				{
					Name:      "terraform",
					Usage:     "lists terraform modules for a repository",
					ArgsUsage: "REPO_ID",
					Action:    p.handleTerraforma,
				},
				{
					Name:      "versions",
					Usage:     "lists versions of a chart",
					ArgsUsage: "CHART_ID",
					Action:    p.handleVersions,
				},
				{
					Name:      "chartinstallations",
					Aliases:   []string{"ci"},
					Usage:     "lists chart installations for a repository",
					ArgsUsage: "REPO_ID",
					Action:    p.handleChartInstallations,
				},
				{
					Name:      "terraforminstallations",
					Aliases:   []string{"ti"},
					Usage:     "lists terraform installations for a repository",
					ArgsUsage: "REPO_ID",
					Action:    p.handleTerraformInstallations,
				},
				{
					Name:      "artifacts",
					Usage:     "Lists artifacts for a repository",
					ArgsUsage: "REPO_ID",
					Action:    p.handleArtifacts,
				},
			},
		},
	}
}

func (p *Plural) handleInstallations(c *cli.Context) error {
	installations, err := p.GetInstallations()
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Repository", "Repository Id", "Publisher"})
	for _, inst := range installations {
		if inst.Repository != nil {
			repo := inst.Repository
			publisherName := ""
			if repo.Publisher != nil {
				publisherName = repo.Publisher.Name
			}
			table.Append([]string{repo.Name, repo.Id, publisherName})
		}
	}
	table.Render()
	return nil
}

func (p *Plural) handleCharts(c *cli.Context) error {
	charts, err := p.GetCharts(c.Args().First())
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Name", "Description", "Latest Version"})
	for _, chart := range charts {
		table.Append([]string{chart.Id, chart.Name, chart.Description, chart.LatestVersion})
	}
	table.Render()
	return nil
}

func (p *Plural) handleTerraforma(c *cli.Context) error {
	tfs, err := p.GetTerraforma(c.Args().First())
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Name", "Description"})
	for _, tf := range tfs {
		table.Append([]string{tf.Id, tf.Name, tf.Description})
	}
	table.Render()
	return nil
}

func (p *Plural) handleVersions(c *cli.Context) error {
	versions, err := p.GetVersions(c.Args().First())

	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Version"})
	for _, version := range versions {
		table.Append([]string{version.Id, version.Version})
	}
	table.Render()
	return nil
}

func (p *Plural) handleChartInstallations(c *cli.Context) error {
	chartInstallations, err := p.GetChartInstallations(c.Args().First())

	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Chart Id", "Chart Name", "Version"})
	for _, ci := range chartInstallations {
		if ci.Chart != nil && ci.Version != nil {
			table.Append([]string{ci.Id, ci.Chart.Id, ci.Chart.Name, ci.Version.Version})
		}
	}
	table.Render()
	return nil
}

func (p *Plural) handleTerraformInstallations(c *cli.Context) error {
	terraformInstallations, err := p.GetTerraformInstallations(c.Args().First())

	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Terraform Id", "Name"})
	for _, ti := range terraformInstallations {
		tf := ti.Terraform
		if tf != nil {
			table.Append([]string{ti.Id, tf.Id, tf.Name})
		}
	}
	table.Render()
	return nil
}

func (p *Plural) handleArtifacts(c *cli.Context) error {
	artifacts, err := p.ListArtifacts(c.Args().First())

	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Id", "Name", "Platform", "Blob", "Sha"})
	for _, artifact := range artifacts {
		table.Append([]string{artifact.Id, artifact.Name, artifact.Platform, artifact.Blob, artifact.Sha})
	}
	table.Render()
	return nil
}
