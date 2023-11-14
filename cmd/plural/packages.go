package plural

import (
	"fmt"
	"os"

	"github.com/pluralsh/plural-cli/pkg/api"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"

	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/wkspace"
)

func (p *Plural) packagesCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "install",
			Usage:     "installs a package at a specific version",
			ArgsUsage: "helm|terraform REPO NAME VSN",
			Action:    affirmed(requireArgs(p.installPackage, []string{"TYPE", "REPO", "NAME", "VERSION"}), "Are you sure you want to install this package?", "PLURAL_PACKAGES_INSTALL"),
		},
		{
			Name:      "uninstall",
			Usage:     "uninstall a helm or terraform package",
			ArgsUsage: "helm|terraform REPO NAME",
			Action:    latestVersion(affirmed(requireArgs(rooted(p.uninstallPackage), []string{"TYPE", "REPO", "NAME"}), "Are you sure you want to uninstall this package?", "PLURAL_PACKAGES_UNINSTALL")),
		},
		{
			Name:      "list",
			Usage:     "lists the packages installed for a given repo",
			ArgsUsage: "REPO",
			Action:    latestVersion(requireArgs(rooted(p.listPackages), []string{"REPO"})),
		},
		{
			Name:        "show",
			Usage:       "Shows version information for packages within a plural repo",
			Subcommands: p.showCommands(),
		},
	}
}

func (p *Plural) showCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "helm",
			Usage:     "list versions for a helm chart",
			ArgsUsage: "REPO NAME",
			Action:    requireArgs(p.showHelm, []string{"REPO", "NAME"}),
		},
		{
			Name:      "terraform",
			Usage:     "list versions for a terraform module",
			ArgsUsage: "REPO NAME",
			Action:    requireArgs(p.showTerraform, []string{"REPO", "NAME"}),
		},
	}
}

func (p *Plural) installPackage(c *cli.Context) error {
	p.InitPluralClient()
	tp, repo, name, vsn := c.Args().Get(0), c.Args().Get(1), c.Args().Get(2), c.Args().Get(3)
	if err := p.Client.InstallVersion(tp, repo, name, vsn); err != nil {
		return err
	}

	utils.Success("Successfully installed %s %s version %s in %s\n", tp, name, vsn, repo)
	utils.Highlight("To apply the module in your cluster, you'll need to run `plural build --only %s && plural deploy", repo)
	return nil
}

func (p *Plural) showHelm(c *cli.Context) error {
	p.InitPluralClient()
	repo, name := c.Args().Get(0), c.Args().Get(1)
	chart, err := api.FindChart(p.Client, repo, name)
	if err != nil {
		return err
	}

	vsns, err := p.Client.GetVersions(chart.Id)
	if err != nil {
		return err
	}

	header := []string{"Name", "Version", "App Version", "Created"}
	return utils.PrintTable(vsns, header, func(vsn *api.Version) ([]string, error) {
		appVsn := ""
		if app, ok := vsn.Helm["appVersion"]; ok {
			if v, ok := app.(string); ok {
				appVsn = v
			}
		}
		return []string{chart.Name, vsn.Version, appVsn, vsn.InsertedAt}, nil
	})
}

func (p *Plural) showTerraform(c *cli.Context) error {
	p.InitPluralClient()
	repo, name := c.Args().Get(0), c.Args().Get(1)
	chart, err := api.FindTerraform(p.Client, repo, name)
	if err != nil {
		return err
	}

	vsns, err := p.Client.GetTerraformVersions(chart.Id)
	if err != nil {
		return err
	}

	header := []string{"Name", "Version", "Created"}
	return utils.PrintTable(vsns, header, func(vsn *api.Version) ([]string, error) {
		return []string{chart.Name, vsn.Version, vsn.InsertedAt}, nil
	})
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
				return api.GetErrorResponse(p.Client.UninstallTerraform(inst.Id), "UninstallTerraform")
			}
		}
	}

	if t == "helm" {
		for _, inst := range space.Charts {
			if inst.Chart.Name == name {
				return api.GetErrorResponse(p.Client.UninstallChart(inst.Id), "UninstallChart")
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
		return nil, api.GetErrorResponse(err, "GetInstallation")
	}

	if inst == nil {
		return nil, fmt.Errorf("no installation found for package: %s", repo)
	}

	return wkspace.New(p.Client, inst)
}
