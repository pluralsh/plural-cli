package main

import (
	"os"

	"github.com/michaeljguarino/forge/proxy"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

func proxyCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "list",
			Usage:     "lists proxy plugins for a repo",
			ArgsUsage: "REPO",
			Action:    handleProxyList,
		},
		{
			Name:      "connect",
			Usage:     "connects to a named proxy for a repo",
			ArgsUsage: "REPO NAME",
			Action:    handleProxyConnect,
		},
	}
}

func handleProxyList(c *cli.Context) error {
	repo := c.Args().Get(0)
	proxies, err := proxy.List(repo)
	if err != nil {
		return err
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Type", "Target"})
	for _, p := range proxies.Items {
		table.Append([]string{p.Name, p.Spec.Type, p.Spec.Target})
	}
	table.Render()
	return nil
}

func handleProxyConnect(c *cli.Context) error {
	repo := c.Args().Get(0)
	name := c.Args().Get(1)
	return proxy.Exec(repo, name)
}
