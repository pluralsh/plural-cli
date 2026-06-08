package agents

import (
	"github.com/urfave/cli"

	"github.com/pluralsh/plural-cli/pkg/client"
)

func init() {
	consoleToken = ""
	consoleURL = ""
}

var consoleToken string
var consoleURL string

type Plural struct {
	client.Plural
	service *Service
}

func Command(clients client.Plural) cli.Command {
	p := &Plural{Plural: clients}
	p.service = NewService(&p.Plural)
	return cli.Command{
		Name:        "agents",
		Usage:       "list and resume plural agent runs",
		Subcommands: p.commands(),
		Category:    "AI",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:        "token",
				Usage:       "console token",
				EnvVar:      "PLURAL_CONSOLE_TOKEN",
				Destination: &consoleToken,
			},
			cli.StringFlag{
				Name:        "url",
				Usage:       "console url address",
				EnvVar:      "PLURAL_CONSOLE_URL",
				Destination: &consoleURL,
			},
		},
	}
}

func (p *Plural) commands() []cli.Command {
	return []cli.Command{
		{
			Name:      "resume",
			Usage:     "restore and resume a plural agent run locally",
			ArgsUsage: "[run-id]",
			Action:    p.handleResume,
		},
	}
}
