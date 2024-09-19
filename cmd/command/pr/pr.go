package pr

import (
	"io"
	"os"

	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/pr"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/samber/lo"
	"github.com/urfave/cli"
)

func init() {
	consoleToken = ""
	consoleURL = ""
}

var consoleToken string
var consoleURL string

type Plural struct {
	client.Plural
}

func Command(clients client.Plural) cli.Command {
	p := Plural{
		Plural: clients,
	}
	return cli.Command{
		Name:        "pull-requests",
		Aliases:     []string{"pr"},
		Usage:       "Generate and manage pull requests",
		Subcommands: p.prCommands(),
		Category:    "CD",
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

func (p *Plural) prCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "template",
			Usage:  "applies a pr template resource in the local source tree",
			Action: handlePrTemplate,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "file",
					Usage:    "the file the template was placed in",
					Required: true,
				},
				cli.StringFlag{
					Name:     "templates",
					Usage:    "a directory of external templates to use for creating new files",
					Required: false,
				},
			},
		},
		{
			Name:      "create",
			Action:    common.LatestVersion(common.RequireArgs(p.handleCreatePrAutomation, []string{"ID"})),
			Usage:     "create PR automation",
			ArgsUsage: "ID",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "context", Usage: "JSON blob string"},
				cli.StringFlag{Name: "branch", Usage: "branch name"},
			},
		},
		{
			Name:   "test",
			Action: common.LatestVersion(handleTestPrAutomation),
			Usage:  "tests a PR automation CRD locally",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "file",
					Usage:    "the file the PR automation was placed in",
					Required: true,
				},
				cli.StringFlag{
					Name:     "context",
					Usage:    "a yaml file containing the context for the PRA, will read from stdin if not present",
					Required: false,
				},
			},
		},
	}
}

func handlePrTemplate(c *cli.Context) error {
	template, err := pr.Build(c.String("file"))
	if err != nil {
		return err
	}

	if template.Spec.Creates != nil {
		template.Spec.Creates.ExternalDir = c.String("templates")
	}

	return pr.Apply(template)
}

func handleTestPrAutomation(c *cli.Context) error {
	template, err := pr.BuildCRD(c.String("file"), c.String("context"))
	if err != nil {
		return err
	}

	if template.Spec.Creates != nil {
		template.Spec.Creates.ExternalDir = c.String("templates")
	}

	return pr.Apply(template)
}

func (p *Plural) handleCreatePrAutomation(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	var branch, context *string
	var prID string
	id, name := common.GetIdAndName(c.Args().Get(0))
	if id != nil {
		prID = *id
	}
	if name != nil {
		pr, err := p.ConsoleClient.GetPrAutomationByName(*name)
		if err != nil {
			return err
		}
		prID = pr.ID
	}

	if c := c.String("context"); c != "" {
		if c == "-" {
			bytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			context = lo.ToPtr(string(bytes))
		}
	}

	if b := c.String("branch"); b != "" {
		branch = &b
	}

	pr, err := p.ConsoleClient.CreatePullRequest(prID, branch, context)
	if err != nil {
		return err
	}

	utils.Success("PR %s created successfully\n", pr.ID)
	return nil
}
