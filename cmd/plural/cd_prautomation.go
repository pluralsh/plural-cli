package plural

import (
	"io"
	"os"

	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/samber/lo"
	"github.com/urfave/cli"
)

func (p *Plural) cdPrAutomations() cli.Command {
	return cli.Command{
		Name:        "pr-automation",
		Subcommands: p.cdPrAutomationsCommands(),
		Usage:       "manage PR automations",
	}
}

func (p *Plural) cdPrAutomationsCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "create",
			Action:    latestVersion(requireArgs(p.handleCreatePrAutomation, []string{"ID"})),
			Usage:     "create PR automation",
			ArgsUsage: "ID",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "context", Usage: "JSON blob string"},
				cli.StringFlag{Name: "branch", Usage: "branch name"},
			},
		},
	}
}

func (p *Plural) handleCreatePrAutomation(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	var branch, context *string

	id := c.Args().Get(0)
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

	pr, err := p.ConsoleClient.CreatePullRequest(id, branch, context)
	if err != nil {
		return err
	}

	utils.Success("PR %s created successfully\n", pr.ID)
	return nil
}
