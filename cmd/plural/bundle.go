package main

import (
	"github.com/olekukonko/tablewriter"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/bundle"
	"github.com/urfave/cli"
	"os"
)

func bundleCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "list",
			Usage:     "lists bundles for a repository",
			ArgsUsage: "[repo]",
			Action:    bundleList,
		},
		{
			Name:      "install",
			Usage:     "installs a bundle and writes the configuration to this installation's context",
			ArgsUsage: "[repo] [name]",
			Action:    bundleInstall,
		},
	}
}

func bundleList(c *cli.Context) error {
	client := api.NewClient()
	recipes, err := client.ListRecipes(c.Args().Get(0))
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Description", "Provider"})
	for _, recipe := range recipes {
		table.Append([]string{recipe.Name, recipe.Description, recipe.Provider})
	}

	table.Render()
	return nil
}

func bundleInstall(c *cli.Context) error {
	args := c.Args()
	return bundle.Install(args.Get(0), args.Get(1))
}
