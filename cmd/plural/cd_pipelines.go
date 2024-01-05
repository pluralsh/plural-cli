package plural

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

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
	fmt.Printf("Pipeline attributes: %+v\n", attrs)
	PrettyPrintStruct(attrs, 3)

	pipe, err := p.ConsoleClient.SavePipeline(name, *attrs)
	if err != nil {
		fmt.Printf("Error saving pipeline: %v\n", err)
		return err
	}

	utils.Success("Pipeline %s created successfully\n", pipe.Name)
	return nil
}

func PrettyPrintStruct(v interface{}, depth int) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		PrettyPrintStruct(rv.Elem().Interface(), depth)
		return
	}

	if rv.Kind() != reflect.Struct {
		fmt.Printf("%s%v\n", strings.Repeat("  ", depth), rv.Interface())
		return
	}

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rv.Type().Field(i)
		fmt.Printf("%s%s: ", strings.Repeat("  ", depth), fieldType.Name)

		if field.Kind() == reflect.Struct {
			fmt.Println()
			PrettyPrintStruct(field.Interface(), depth+1)
		} else if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				fmt.Println("nil")
			} else {
				PrettyPrintStruct(field.Elem().Interface(), depth+1)
			}
		} else if field.Kind() == reflect.Slice || field.Kind() == reflect.Array {
			fmt.Println()
			for j := 0; j < field.Len(); j++ {
				PrettyPrintStruct(field.Index(j).Interface(), depth+1)
			}
		} else {
			fmt.Println(field.Interface())
		}
	}
}
