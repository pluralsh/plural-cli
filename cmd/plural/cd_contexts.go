package plural

import (
	"encoding/json"
	"fmt"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/urfave/cli"
)

func (p *Plural) cdContexts() cli.Command {
	return cli.Command{
		Name:        "contexts",
		Subcommands: p.cdServiceContextCommands(),
		Usage:       "manage CD service contexts",
	}
}

func (p *Plural) cdServiceContextCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "upsert",
			ArgsUsage: "NAME",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "config-file", Usage: "path for json configuration file with the context blob", Required: true},
				cli.StringFlag{Name: "name", Usage: "context name", Required: true},
			},
			Action: latestVersion(requireArgs(p.handleUpsertServiceContext, []string{"NAME"})),
			Usage:  "upsert service context",
		},
		{
			Name:      "get",
			ArgsUsage: "NAME",
			Action:    latestVersion(requireArgs(p.handleGetServiceContext, []string{"NAME"})),
			Usage:     "get service context",
		},
	}
}

func (p *Plural) handleUpsertServiceContext(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	contextName := c.String("name")
	serviceContextName := c.Args().Get(0)
	attributes := gqlclient.ServiceContextAttributes{}

	configFile, err := utils.ReadFile(c.String("config-file"))
	if err != nil {
		return err
	}

	// validate
	conf := map[string]interface{}{}
	if err := json.Unmarshal([]byte(configFile), &conf); err != nil {
		return err
	}

	configuration := map[string]map[string]interface{}{}
	configuration[contextName] = conf

	configurationJson, err := json.Marshal(configuration)
	if err != nil {
		return err
	}
	configurationJsonString := string(configurationJson)
	attributes.Configuration = &configurationJsonString

	sc, err := p.ConsoleClient.SaveServiceContext(serviceContextName, attributes)
	if err != nil {
		return err
	}
	if sc == nil {
		return fmt.Errorf("the returned object is empty, check if all fields are set")
	}

	desc, err := console.DescribeServiceContext(sc)
	if err != nil {
		return err
	}
	fmt.Print(desc)
	return nil
}

func (p *Plural) handleGetServiceContext(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	contextName := c.Args().Get(0)

	sc, err := p.ConsoleClient.GetServiceContext(contextName)
	if err != nil {
		return err
	}
	if sc == nil {
		return fmt.Errorf("the returned object is empty, check if all fields are set")
	}

	desc, err := console.DescribeServiceContext(sc)
	if err != nil {
		return err
	}
	fmt.Print(desc)
	return nil
}
