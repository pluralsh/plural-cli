package api

import (
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/urfave/cli"
)

type Plural struct {
	client.Plural
}

func Command(clients client.Plural) cli.Command {
	plural := Plural{
		Plural: clients,
	}
	return cli.Command{
		Name:        "api",
		Usage:       "inspect the plural api",
		Subcommands: plural.apiCommands(),
		Category:    "API",
	}
}

func (p *Plural) apiCommands() []cli.Command {
	return []cli.Command{
		{
			Name:  "create",
			Usage: "creates plural resources",
			Subcommands: []cli.Command{
				{
					Name:      "domain",
					Usage:     "creates a new domain for your account",
					ArgsUsage: "{domain}",
					Action:    common.LatestVersion(common.RequireArgs(p.handleCreateDomain, []string{"{domain}"})),
				},
			},
		},
	}
}

func (p *Plural) handleCreateDomain(c *cli.Context) error {
	p.InitPluralClient()
	err := p.CreateDomain(c.Args().First())
	return api.GetErrorResponse(err, "CreateDomain")
}
