package main

import (
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/wkspace"
	"github.com/urfave/cli"
)

func (p *Plural) packagesCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "uninstall",
			Usage:     "uninstall a helm or terraform package",
			ArgsUsage: "TYPE REPO NAME",
			Action:    affirmed(requireArgs(rooted(p.uninstallPackage), []string{"TYPE", "REPO", "NAME"}), "Are you sure you want to uninstall this package?"),
		},
		{
			Name:      "list",
			Usage:     "lists the packages installed for a given repo",
			ArgsUsage: "REPO",
			Action:    requireArgs(rooted(p.listPackages), []string{"REPO"}),
		},
	}
}

func (p *Plural) listPackages(c *cli.Context) error {
	p.InitPluralClient()
	repo := c.Args().Get(0)
	space, err := p.getWorkspace(repo)
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Type", "Name", "Version"})
	for _, inst := range space.Terraform {
		table.Append([]string{"terraform", inst.Terraform.Name, inst.Version.Version})
	}

	for _, inst := range space.Charts {
		table.Append([]string{"helm", inst.Chart.Name, inst.Version.Version})
	}

	table.Render()
	return nil
}

func (p *Plural) uninstallPackage(c *cli.Context) error {
	p.InitPluralClient()
	args := c.Args()
	t, repo, name := args.Get(0), args.Get(1), args.Get(2)

	space, err := p.getWorkspace(repo)
	if err != nil {
		return err
	}

	if t == "terraform" {
		for _, inst := range space.Terraform {
			if inst.Terraform.Name == name {
				return p.Client.UninstallTerraform(inst.Id)
			}
		}
	}

	if t == "helm" {
		for _, inst := range space.Charts {
			if inst.Chart.Name == name {
				return p.Client.UninstallChart(inst.Id)
			}
		}
	}

	utils.Warn("Could not find %s package %s in %s", t, name, repo)
	return nil
}

func (p *Plural) getWorkspace(repo string) (*wkspace.Workspace, error) {
	p.InitPluralClient()
	inst, err := p.Client.GetInstallation(repo)
	if err != nil {
		return nil, err
	}

	return wkspace.New(p.Client, inst)
}
