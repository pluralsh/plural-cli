package main

import (
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/logs"
	"github.com/urfave/cli/v2"
)

func (p *Plural) logsCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:      "list",
			Usage:     "lists log tails for a repo",
			ArgsUsage: "REPO",
			Action:    latestVersion(requireArgs(p.handleLogsList, []string{"REPO"})),
		},
		{
			Name:      "tail",
			Usage:     "execs the specific logtail",
			ArgsUsage: "REPO NAME",
			Action:    latestVersion(requireArgs(p.handleLogTail, []string{"REPO", "NAME"})),
		},
	}
}

func (p *Plural) handleLogsList(c *cli.Context) error {
	repo := c.Args().Get(0)
	conf := config.Read()
	if err := p.InitKube(); err != nil {
		return err
	}
	tails, err := logs.List(p.Kube, conf.Namespace(repo))
	if err != nil {
		return err
	}

	return logs.Print(tails)
}

func (p *Plural) handleLogTail(c *cli.Context) error {
	repo := c.Args().Get(0)
	name := c.Args().Get(1)
	conf := config.Read()
	if err := p.InitKube(); err != nil {
		return err
	}
	return logs.Tail(p.Kube, conf.Namespace(repo), name)
}
