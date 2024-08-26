package cd

import (
	"io"
	"os"

	"github.com/pluralsh/plural-cli/pkg/common"

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
			Action: common.LatestVersion(common.RequireArgs(p.handleCreatePipeline, []string{})),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "file",
					Usage: "the file this pipeline is defined in, use - for stdin",
				},
			},
		},
		{
			Name:      "context",
			Action:    common.LatestVersion(common.RequireArgs(p.handlePipelineContext, []string{"PIPELINE_NAME"})),
			Usage:     "update pipeline context",
			ArgsUsage: "PIPELINE_NAME",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:     "set",
					Usage:    "key-value pairs to put in the context, i.e. key.subkey=value",
					Required: true,
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

	utils.Success("Pipeline %s created successfully\n", pipe.Name)
	return nil
}

func (p *Plural) handlePipelineContext(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	pipelineName := c.Args().Get(0)

	var setArgs []string
	if c.IsSet("set") {
		setArgs = append(setArgs, c.StringSlice("set")...)
	}

	// TODO

	utils.Success("Pipeline %s updated successfully\n", pipelineName)
	return nil
}
