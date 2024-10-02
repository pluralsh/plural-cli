package ops

import (
	"fmt"

	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/provider"
	"github.com/pluralsh/plural-cli/pkg/utils"
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
		Name:        "ops",
		Usage:       "Commands for simplifying cluster operations",
		Subcommands: p.opsCommands(),
		Category:    "Debugging",
	}
}

func (p *Plural) opsCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "terminate",
			Usage:     "terminates a worker node in your cluster",
			ArgsUsage: "{name}",
			Action:    common.LatestVersion(common.RequireArgs(common.InitKubeconfig(p.handleTerminateNode), []string{"{name}"})),
		},
		{
			Name:   "cluster",
			Usage:  "list the nodes in your cluster",
			Action: common.LatestVersion(common.InitKubeconfig(p.handleListNodes)),
		},
	}
}

func (p *Plural) handleListNodes(c *cli.Context) error {
	if err := p.InitKube(); err != nil {
		return err
	}
	nodes, err := p.Nodes()
	if err != nil {
		return err
	}
	return common.PrintListNodes(nodes)
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

func getProvider() (provider.Provider, error) {
	_, found := utils.ProjectRoot()
	if !found {
		return nil, fmt.Errorf("project not initialized, run `plural init` to set up a workspace")
	}

	return provider.GetProvider()
}
