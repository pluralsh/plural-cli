package stack

import (
	"fmt"

	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"

	"github.com/urfave/cli"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/bundle"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

type Plural struct {
	client.Plural
}

func Command(clients client.Plural) cli.Command {
	p := Plural{
		Plural: clients,
	}
	return cli.Command{
		Name:        "stack",
		Usage:       "Commands for installing and discovering plural stacks",
		Subcommands: p.stackCommands(),
	}
}

func (p *Plural) stackCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "install",
			Usage:     "installs a plural stack for your current provider",
			ArgsUsage: "{stack-name}",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "refresh",
					Usage: "re-enter the configuration for all bundles",
				},
			},
			Action: common.Tracked(common.LatestVersion(common.Rooted(common.RequireArgs(p.stackInstall, []string{"{stack-name}"}))), "stack.install"),
		},
		{
			Name:  "list",
			Usage: "lists stacks to potentially install",
			Flags: []cli.Flag{
				cli.BoolTFlag{
					Name:  "account",
					Usage: "only list stacks within your account",
				},
			},
			Action: common.LatestVersion(common.Rooted(p.stackList)),
		},
	}
}

func (p *Plural) stackInstall(c *cli.Context) (err error) {
	name := c.Args().Get(0)
	man, err := manifest.FetchProject()
	if err != nil {
		return
	}

	p.InitPluralClient()
	err = bundle.Stack(p.Client, name, man.Provider, c.Bool("refresh"))
	utils.Note("To edit the configuration you've just entered, edit the context.yaml file at the root of your repo, or run with the --refresh flag\n")
	return
}

func (p *Plural) stackList(c *cli.Context) (err error) {
	p.InitPluralClient()
	stacks, err := p.ListStacks(c.Bool("account"))
	if err != nil {
		return api.GetErrorResponse(err, "ListStacks")
	}

	headers := []string{"Name", "Description", "Featured"}
	return utils.PrintTable(stacks, headers, func(s *api.Stack) ([]string, error) {
		return []string{s.Name, s.Description, fmt.Sprintf("%v", s.Featured)}, nil
	})
}
