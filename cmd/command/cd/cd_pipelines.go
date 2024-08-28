package cd

import (
	"encoding/json"
	"io"
	"os"

	"github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"k8s.io/helm/pkg/strvals"

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
			Action:    common.LatestVersion(common.RequireArgs(p.handlePipelineContext, []string{"PIPELINE_ID"})),
			Usage:     "set pipeline context",
			ArgsUsage: "PIPELINE_ID",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:     "set",
					Usage:    "key-value pairs to put in the context, dot notation is supported, i.e. key.subkey=value",
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

	var setArgs []string
	if c.IsSet("set") {
		setArgs = append(setArgs, c.StringSlice("set")...)
	}

	context := map[string]any{}
	for _, arg := range setArgs {
		if err := strvals.ParseInto(arg, context); err != nil {
			return err
		}
	}

	data, err := json.Marshal(context)
	if err != nil {
		return err
	}

	id := c.Args().Get(0)
	attrs := client.PipelineContextAttributes{Context: string(data)}
	_, err = p.ConsoleClient.CreatePipelineContext(id, attrs)
	if err != nil {
		return err
	}

	utils.Success("Pipeline %s context set successfully\n", id)
	return nil
}
