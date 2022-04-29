package main

import (
	"os"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/urfave/cli"
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
			Name:      "reset",
			Usage:     "eliminates your current plural installation set, to change cloud provider or eject from plural",
			Action:    handleResetInstallations,
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

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Description", "Publisher"})
	for _, repo := range repos {
		table.Append([]string{repo.Name, repo.Description, repo.Publisher.Name})
	}

	table.Render()
	return nil
}

func handleResetInstallations(c *cli.Context) error {
	client := api.NewClient()
	count, err := client.ResetInstallations()
	if err != nil {
		return err
	}

	fmt.Printf("Deleted %d installations in app.plural.sh\n", count)
	fmt.Println("(you can recreate these at any time and any running infrastructure is not affected, plural will simply no longer deliver upgrades)")
	return nil
}