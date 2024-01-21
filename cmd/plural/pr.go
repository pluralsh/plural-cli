package plural

import (
	"github.com/pluralsh/plural-cli/pkg/pr"
	"github.com/urfave/cli"
)

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
			},
		},
	}
}

func handlePrTemplate(c *cli.Context) error {
	template, err := pr.Build(c.String("file"))
	if err != nil {
		return err
	}

	return pr.Apply(template)
}
