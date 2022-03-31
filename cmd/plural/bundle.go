package main

import (
	"github.com/olekukonko/tablewriter"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/bundle"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/urfave/cli"
	"os"
	"strings"
)

func bundleCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "list",
			Usage:     "lists bundles for a repository",
			ArgsUsage: "[repo]",
			Action:    requireArgs(bundleList, []string{"repo"}),
		},
		{
			Name:      "install",
			Usage:     "installs a bundle and writes the configuration to this installation's context",
			ArgsUsage: "[repo] [name]",
			Flags:     []cli.Flag{
				cli.BoolFlag{
					Name: "refresh",
					Usage: "re-enter the configuration for this bundle",
				},
			},
			Action:    requireArgs(bundleInstall, []string{"repo", "bundle-name"}),
		},
	}
}

func bundleList(c *cli.Context) error {
	client := api.NewClient()
	man, err := manifest.FetchProject()
	if err != nil {
		return err
	} 

	recipes, err := client.ListRecipes(c.Args().Get(0), strings.ToUpper(man.Provider))
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

func bundleInstall(c *cli.Context) (err error) {
	args := c.Args()
	err = bundle.Install(args.Get(0), args.Get(1), c.Bool("refresh"))
	utils.Note("To edit the configuration you've just entered, edit the context.yaml file at the root of your repo, or run with the --refresh flag\n")
	return
}
