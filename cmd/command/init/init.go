package init

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
		Name:  "init",
		Usage: "initializes plural within a git repo",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "endpoint",
				Usage: "the endpoint for the plural installation you're working with",
			},
			cli.StringFlag{
				Name:  "service-account",
				Usage: "email for the service account you'd like to use for this workspace",
			},
			cli.BoolFlag{
				Name:  "ignore-preflights",
				Usage: "whether to ignore preflight check failures prior to init",
			},
		},
		Action: common.Tracked(common.LatestVersion(p.HandleInit), "cli.init"),
	}
}
