package plural

import (
	"encoding/json"
	"io"
	"os"

	gqlclient "github.com/pluralsh/console-client-go"
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
				cli.StringFlag{Name: "context", Usage: "JSON blob string", Required: true},
			},
		},
	}
}

func (p *Plural) handleCreatePrAutomation(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	name := c.Args().Get(0)
	context := c.String("context")
	if context == "-" {
		bytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		context = string(bytes)
	}
	attrs := &gqlclient.PrAutomationAttributes{}

	err := json.Unmarshal([]byte(context), attrs)
	if err != nil {
		return err
	}
	attrs.Name = lo.ToPtr(name)
	_, err = p.ConsoleClient.CreatePrAutomation(*attrs)
	if err != nil {
		return err
	}

	utils.Success("PR automation %s created successfully\n", name)
	return nil
}
