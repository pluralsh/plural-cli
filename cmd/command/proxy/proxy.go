package proxy

import (
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/proxy"
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
		Name:        "proxy",
		Usage:       "proxies into running processes in your cluster",
		Subcommands: p.proxyCommands(),
		Category:    "Debugging",
	}
}

func (p *Plural) proxyCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "list",
			Usage:     "lists proxy plugins for a repo",
			ArgsUsage: "{repo}",
			Action:    common.LatestVersion(common.InitKubeconfig(common.RequireArgs(p.handleProxyList, []string{"{repo}"}))),
		},
		{
			Name:      "connect",
			Usage:     "connects to a named proxy for a repo",
			ArgsUsage: "{repo} {name}",
			Action:    common.LatestVersion(common.InitKubeconfig(common.RequireArgs(p.handleProxyConnect, []string{"{repo}", "{name}"}))),
		},
	}
}

func (p *Plural) handleProxyList(c *cli.Context) error {
	repo := c.Args().Get(0)
	conf := config.Read()
	if err := p.InitKube(); err != nil {
		return err
	}
	proxies, err := proxy.List(p.Kube, conf.Namespace(repo))
	if err != nil {
		return err
	}

	return proxy.Print(proxies)
}

func (p *Plural) handleProxyConnect(c *cli.Context) error {
	repo := c.Args().Get(0)
	name := c.Args().Get(1)
	conf := config.Read()
	if err := p.InitKube(); err != nil {
		return err
	}
	return proxy.Exec(p.Kube, conf.Namespace(repo), name)
}
