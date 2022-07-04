package main

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/format"
	"github.com/pluralsh/plural/pkg/utils"
)

func reposCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "unlock",
			Usage:     "unlocks installations in a repo that have breaking changes",
			ArgsUsage: "REPO",
			Action:    handleUnlockRepo,
		},
		{
			Name:   "reset",
			Usage:  "eliminates your current plural installation set, to change cloud provider or eject from plural",
			Action: handleResetInstallations,
		},
		{
			Name:      "list",
			Usage:     "list available repositories to install",
			ArgsUsage: "",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "query",
					Usage: "string to search by",
				},
				cli.StringFlag{
					Name:  "format",
					Usage: "format to print the repositories out, eg csv or default is table",
				},
			},
			Action: handleListRepositories,
		},
	}
}

func handleUnlockRepo(c *cli.Context) error {
	client := api.NewClient()
	return client.UnlockRepository(c.Args().First())
}

func handleListRepositories(c *cli.Context) error {
	client := api.NewClient()
	repos, err := client.ListRepositories(c.String("query"))
	if err != nil {
		return err
	}

	addIcon := c.String("format") == "csv"

	formatter := format.New(format.FormatType(c.String("format")))
	header := []string{"Repo", "Description", "Publisher", "Bundles"}
	if addIcon {
		header = append(header, "Icon")
	}

	formatter.Header(header)
	for _, repo := range repos {
		recipeNames := utils.Map(repo.Recipes, func(recipe *api.Recipe) string {
			return recipe.Name
		})

		line := []string{repo.Name, repo.Description, repo.Publisher.Name, strings.Join(recipeNames, ", ")}
		if addIcon {
			line = append(line, repo.Icon)
		}
		err := formatter.Write(line)
		if err != nil {
			return err
		}
	}

	if err := formatter.Flush(); err != nil {
		return err
	}
	return nil
}

func handleResetInstallations(*cli.Context) error {
	client := api.NewClient()
	count, err := client.ResetInstallations()
	if err != nil {
		return err
	}

	fmt.Printf("Deleted %d installations in app.plural.sh\n", count)
	fmt.Println("(you can recreate these at any time and any running infrastructure is not affected, plural will simply no longer deliver upgrades)")
	return nil
}
