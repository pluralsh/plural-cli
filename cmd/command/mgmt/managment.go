package mgmt

import (
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
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
		Name:        "management",
		Aliases:     []string{"mgmt"},
		Usage:       "manages installations in your workspace",
		Subcommands: p.managementCommands(),
		Category:    "Workspace",
	}
}

func (p *Plural) managementCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "kube-init",
			Usage:  "generates kubernetes credentials for this subworkspace",
			Action: common.LatestVersion(common.KubeInit),
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
