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
			Action:    latestVersion(requireArgs(p.handleCreatePrAutomation, []string{"NAME"})),
			Usage:     "create PR automation",
			ArgsUsage: "NAME",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "context", Usage: "JSON blob string"},
				cli.StringFlag{Name: "branch", Usage: "branch name"},
				cli.StringFlag{Name: "automation-id", Usage: "the ID of the PR automation", Required: true},
			},
		},
	}
}

func (p *Plural) handleCreatePrAutomation(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	var branch *string
	name := c.Args().Get(0)
	id := c.String("automation-id")
	context := c.String("context")
	if context == "-" {
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		context = string(bytes)
	}
	if b := c.String("branch"); b != "" {
		branch = &b
	}

	_, err := p.ConsoleClient.CreatePullRequest(id, name, branch, lo.ToPtr(context))
	if err != nil {
		return err
	}

	utils.Success("PR %s created successfully\n", name)
	return nil
}
