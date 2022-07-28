package main

import (
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/proxy"
	"github.com/urfave/cli"
)

func (p *Plural) proxyCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "list",
			Usage:     "lists proxy plugins for a repo",
			ArgsUsage: "REPO",
			Action:    requireArgs(p.handleProxyList, []string{"REPO"}),
		},
		{
			Name:      "connect",
			Usage:     "connects to a named proxy for a repo",
			ArgsUsage: "REPO NAME",
			Action:    requireArgs(p.handleProxyConnect, []string{"REPO", "NAME"}),
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

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Type", "Target"})
	for _, p := range proxies.Items {
		table.Append([]string{p.Name, string(p.Spec.Type), p.Spec.Target})
	}
	table.Render()
	return nil
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
