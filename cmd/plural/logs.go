package main

import (
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/logs"
	"github.com/urfave/cli"
)

func logsCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "list",
			Usage:     "lists log tails for a repo",
			ArgsUsage: "REPO",
			Action:    requireArgs(handleLogsList, []string{"REPO"}),
		},
		{
			Name:      "tail",
			Usage:     "execs the specific logtail",
			ArgsUsage: "REPO NAME",
			Action:    requireArgs(handleLogTail, []string{"REPO", "NAME"}),
		},
	}
}

func handleLogsList(c *cli.Context) error {
	repo := c.Args().Get(0)
	conf := config.Read()
	tails, err := logs.List(conf.Namespace(repo))
	if err != nil {
		return err
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Follow", "Target"})
	for _, t := range tails.Items {
		follow := "False"
		if t.Spec.Follow {
			follow = "True"
		}

		table.Append([]string{t.Name, follow, t.Spec.Target})
	}
	table.Render()
	return nil
}

func handleLogTail(c *cli.Context) error {
	repo := c.Args().Get(0)
	name := c.Args().Get(1)
	conf := config.Read()

	return logs.Tail(conf.Namespace(repo), name)
}
