package plural

import (
	"fmt"

	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/urfave/cli"

	gqlclient "github.com/pluralsh/console-client-go"
)

func (p *Plural) cdPipelineGates() cli.Command {
	return cli.Command{
		Name:        "gates",
		Subcommands: p.pipelineGateCommands(),
		Usage:       "manage CD PipelineGates",
	}
}

func (p *Plural) pipelineGateCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "list",
			Action: latestVersion(p.handleListPipelineGates),
			Usage:  "list pipeline gates",
		},
	}
}

func (p *Plural) handleListPipelineGates(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		fmt.Printf("Error initializing client: %v\n", err)
		return err
	}
	pipelineGates, err := p.ConsoleClient.ListPipelineGates()
	if err != nil {
		fmt.Printf("Error getting PipelineGates: %v\n", err)
		return err
	}
	if pipelineGates == nil {
		return fmt.Errorf("returned objects list [ListPipelineGates] is nil")
	}
	headers := []string{"Id", "Name", "Type", "State"}
	return utils.PrintTable(pipelineGates.ClusterGates, headers, func(pgf *gqlclient.PipelineGateFragment) ([]string, error) {
		return []string{pgf.ID, pgf.Name, string(pgf.Type), string(pgf.State)}, nil
	})
}
