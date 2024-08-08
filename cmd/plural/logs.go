package plural

import (
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/logs"
	"github.com/urfave/cli"
)

func (p *Plural) logsCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "list",
			Usage:     "lists log tails for a repo",
			ArgsUsage: "REPO",
			Action:    common.LatestVersion(initKubeconfig(requireArgs(p.handleLogsList, []string{"REPO"}))),
		},
		{
			Name:      "tail",
			Usage:     "execs the specific logtail",
			ArgsUsage: "REPO NAME",
			Action:    common.LatestVersion(initKubeconfig(requireArgs(p.handleLogTail, []string{"REPO", "NAME"}))),
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
