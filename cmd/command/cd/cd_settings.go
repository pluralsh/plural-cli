package cd

import (
	consoleclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/urfave/cli"
)

func (p *Plural) cdSettings() cli.Command {
	return cli.Command{
		Name:        "settings",
		Subcommands: p.cdSettingsCommands(),
		Usage:       "manage CD settings",
	}
}

func (p *Plural) cdSettingsCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "agents",
			ArgsUsage: "{file-path}",
			Action:    common.LatestVersion(common.RequireArgs(p.handleUpdateAgents, []string{"{file-path}"})),
			Usage:     "update agents settings",
		},
	}
}

func (p *Plural) handleUpdateAgents(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	filepath := c.Args().Get(0)
	content, err := utils.ReadFile(filepath)
	if err != nil {
		return err
	}
	attr := &consoleclient.DeploymentSettingsAttributes{
		AgentHelmValues: &content,
	}
	res, err := p.ConsoleClient.UpdateDeploymentSettings(*attr)
	if err != nil {
		return err
	}

	utils.Success("%s settings updated successfully", res.UpdateDeploymentSettings.Name)
	return nil
}
