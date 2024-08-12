package deploy

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
		Name:      "deploy",
		Aliases:   []string{"d"},
		Usage:     "Deploys the current workspace. This command will first sniff out git diffs in workspaces, topsort them, then apply all changes.",
		ArgsUsage: "Workspace",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "silence",
				Usage: "don't display notes for deployed apps",
			},
			cli.BoolFlag{
				Name:  "verbose",
				Usage: "show all command output during execution",
			},
			cli.BoolFlag{
				Name:  "ignore-console",
				Usage: "don't deploy the plural console",
			},
			cli.BoolFlag{
				Name:  "all",
				Usage: "deploy all repos irregardless of changes",
			},
			cli.StringFlag{
				Name:  "commit",
				Usage: "commits your changes with this message",
			},
			cli.StringSliceFlag{
				Name:  "from",
				Usage: "deploys only this application and its dependencies",
			},
			cli.BoolFlag{
				Name:  "force",
				Usage: "use force push when pushing to git",
			},
		},
		Action: common.Tracked(common.LatestVersion(common.Owned(common.Rooted(p.Deploy))), "cli.deploy"),
	}
}
