package main

import (
	"fmt"

	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/urfave/cli"
	v1 "k8s.io/api/core/v1"
)

func (p *Plural) opsCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "terminate",
			Usage:     "terminates a worker node in your cluster",
			ArgsUsage: "NAME",
			Action:    p.handleTerminateNode,
		},
		{
			Name:   "cluster",
			Usage:  "list the nodes in your cluster",
			Action: p.handleListNodes,
		},
	}
}

func (p *Plural) handleTerminateNode(c *cli.Context) error {
	name := c.Args().Get(0)
	provider, err := getProvider()
	if err != nil {
		return err
	}
	if err := p.InitKube(); err != nil {
		return err
	}
	node, err := p.Node(name)
	if err != nil {
		return err
	}

	return provider.Decommision(node)
}

func (p *Plural) handleListNodes(cli *cli.Context) error {
	if err := p.InitKube(); err != nil {
		return err
	}
	nodes, err := p.Nodes()
	if err != nil {
		return err
	}

	headers := []string{"Name", "CPU", "Memory", "Region", "Zone"}
	return utils.PrintTable(nodes.Items, headers, func(node v1.Node) ([]string, error) {
		status := node.Status
		labels := node.ObjectMeta.Labels
		cpu, mem := status.Capacity["cpu"], status.Capacity["memory"]
		return []string{
			node.Name,
			cpu.String(),
			mem.String(),
			labels["topology.kubernetes.io/region"],
			labels["topology.kubernetes.io/zone"],
		}, nil
	})
}

func getProvider() (provider.Provider, error) {
	_, found := utils.ProjectRoot()
	if !found {
		return nil, fmt.Errorf("project not initialized, run `plural init` to set up a workspace")
	}

	return provider.GetProvider()
}
