package plural

import (
	"fmt"
	"io"
	"os"

	"encoding/json"

	gqlclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/urfave/cli"
)

func (p *Plural) cdPipelines() cli.Command {
	return cli.Command{
		Name:        "pipelines",
		Subcommands: p.pipelineCommands(),
		Usage:       "manage CD pipelines",
	}
}

func (p *Plural) pipelineCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "create",
			Action: latestVersion(requireArgs(p.handleCreatePipeline, []string{})),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "file",
					Usage: "the file this pipeline is defined in, use - for stdin",
				},
			},
		},
		{
			Name:      "delete",
			ArgsUsage: "PIPELINE_ID",
			Action:    latestVersion(requireArgs(p.handleDeletePipeline, []string{"PIPELINE_ID"})),
			Flags:     []cli.Flag{},
			Usage:     "delete pipeline with id",
		},
		//{
		//	Name:      "list",
		//	ArgsUsage: "CLUSTER_ID",
		//	Action:    latestVersion(requireArgs(p.handleListPipelines, []string{"CLUSTER_ID"})),
		//	Usage:     "list cluster services",
		//},
		{
			Name:   "list",
			Action: latestVersion(p.handleListPipelines),
			Usage:  "list pipelines",
		},
	}
}

func (p *Plural) handleCreatePipeline(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	var bytes []byte
	var err error
	file := c.String("file")
	if file == "-" {
		bytes, err = io.ReadAll(os.Stdin)
	} else {
		bytes, err = os.ReadFile(file)
	}

	if err != nil {
		return err
	}

	name, attrs, err := console.ConstructPipelineInput(bytes)
	if err != nil {
		fmt.Printf("Error constructing pipeline input: %v\n", err)
		return err
	}

	fmt.Printf("Pipeline name: %s\n", name)
	attrsJSON, err := json.MarshalIndent(attrs, "", "  ")
	if err != nil {
		fmt.Printf("failed to marshalindent pipeline input attributes:\n %s \n", err)
	}
	fmt.Printf("pipeline json from API: \n %s\n", string(attrsJSON))

	pipe, err := p.ConsoleClient.SavePipeline(name, *attrs)
	if err != nil {
		fmt.Printf("Error saving pipeline: %v\n", err)
		return err
	}
	pipeJSON, err := json.MarshalIndent(pipe, "", "  ")
	if err != nil {
		fmt.Printf("failed to marshalindent pipeline input attributes:\n %s \n", err)
	}
	fmt.Printf("pipeline json from API: \n %s\n", string(pipeJSON))

	utils.Success("Pipeline %s created successfully\n", pipe.Name)
	return nil
}

func (p *Plural) handleDeletePipeline(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	pipelineId := c.Args().Get(0)

	pipe, err := p.ConsoleClient.DeletePipeline(pipelineId)
	if err != nil {
		fmt.Printf("Error deleting pipeline: %v\n", err)
		return err
	}

	utils.Success("Pipeline %s with id %s deleted successfully\n", pipe.Name, pipe.ID)
	return nil
}

func (p *Plural) handleListPipelines(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		fmt.Printf("Error initializing client: %v\n", err)
		return err
	}
	pipelines, err := p.ConsoleClient.ListPipelines()
	if err != nil {
		fmt.Printf("Error getting pipelines: %v\n", err)
		return err
	}
	if pipelines == nil {
		return fmt.Errorf("returned objects list [ListPipelines] is nil")
	}
	headers := []string{"Id", "Name"}
	return utils.PrintTable(pipelines.Pipelines.Edges, headers, func(pef *gqlclient.PipelineEdgeFragment) ([]string, error) {
		return []string{pef.Node.ID, pef.Node.Name}, nil
	})
}
