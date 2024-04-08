package plural

import (
	"os"

	consoleclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/urfave/cli"
	"sigs.k8s.io/yaml"
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
			ArgsUsage: "FILENAME",
			Action:    latestVersion(requireArgs(p.handleUpdateAgents, []string{"FILENAME"})),
			Usage:     "update agents settings",
		},
	}
}

func (p *Plural) handleUpdateAgents(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	filepath := c.Args().Get(0)
	content, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	attr := &consoleclient.DeploymentSettingsAttributes{}
	if err := yaml.Unmarshal(content, attr); err != nil {
		return err
	}

	res, err := p.ConsoleClient.UpdateDeploymentSettings(*attr)
	if err != nil {
		return err
	}

	utils.Success("%s settings updated successfully", res.UpdateDeploymentSettings.Name)
	return nil
}
