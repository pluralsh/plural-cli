package log

import (
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/logs"
	"github.com/urfave/cli"
)

type Plural struct {
	client.Plural
}

func Command(clients client.Plural) cli.Command {
	p := Plural{
		Plural: clients,
	}
	return cli.Command{
		Name:        "logs",
		Usage:       "Commands for tailing logs for specific apps",
		Subcommands: p.logsCommands(),
		Category:    "Debugging",
	}
}

func (p *Plural) logsCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "list",
			Usage:     "lists log tails for a repo",
			ArgsUsage: "{repo}",
			Action:    common.LatestVersion(common.InitKubeconfig(common.RequireArgs(p.handleLogsList, []string{"{repo}"}))),
		},
		{
			Name:      "tail",
			Usage:     "execs the specific logtail",
			ArgsUsage: "{repo} {name}",
			Action:    common.LatestVersion(common.InitKubeconfig(common.RequireArgs(p.handleLogTail, []string{"{repo}", "{name}"}))),
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
