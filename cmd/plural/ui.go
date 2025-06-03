//go:build ui || generate

package plural

import (
	"github.com/urfave/cli"

	"github.com/pluralsh/plural/pkg/ui"
)

func (p *Plural) uiCommands() cli.Command {
	return cli.Command{
		Name:   "install",
		Usage:  "opens installer UI that simplifies application configuration",
		Action: p.run,
	}
}

func (p *Plural) run(c *cli.Context) error {
	p.InitPluralClient()
	return ui.Run(p.Client, c)
}
