package pr

import (
	"github.com/pluralsh/plural-cli/pkg/pr"
	"github.com/urfave/cli"
)

func Command() cli.Command {
	return cli.Command{
		Name:        "pull-requests",
		Aliases:     []string{"pr"},
		Usage:       "Generate and manage pull requests",
		Subcommands: prCommands(),
		Category:    "CD",
	}
}

func prCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "template",
			Usage:  "applies a pr template resource in the local source tree",
			Action: handlePrTemplate,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "file",
					Usage:    "the file the template was placed in",
					Required: true,
				},
				cli.StringFlag{
					Name:     "templates",
					Usage:    "a directory of external templates to use for creating new files",
					Required: false,
				},
			},
		},
	}
}

func handlePrTemplate(c *cli.Context) error {
	template, err := pr.Build(c.String("file"))
	if err != nil {
		return err
	}

	if template.Spec.Creates != nil {
		template.Spec.Creates.ExternalDir = c.String("templates")
	}

	return pr.Apply(template)
}
