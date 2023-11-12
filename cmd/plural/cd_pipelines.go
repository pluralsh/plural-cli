package plural

import (
	"io"
	"os"

	"github.com/pluralsh/plural/pkg/console"
	"github.com/pluralsh/plural/pkg/utils"
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
		return err
	}

	pipe, err := p.ConsoleClient.SavePipeline(name, *attrs)
	if err != nil {
		return err
	}

	utils.Success("Pipeline %s created successfully", pipe.Name)
	return nil
}
