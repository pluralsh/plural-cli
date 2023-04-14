//go:build ui || generate

package plural

import (
	"github.com/urfave/cli"

	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/ui"
	"github.com/pluralsh/plural/pkg/wkspace"
)

func (p *Plural) uiCommands() cli.Command {
	return cli.Command{
		Name:   "install",
		Usage:  "opens installer UI that simplifies application configuration",
		Action: tracked(rooted(p.run), "cli.install"),
	}
}

func (p *Plural) run(c *cli.Context) error {
	_, err := wkspace.Preflight()
	if err != nil {
		return err
	}

	_, err = manifest.FetchProject()
	if err != nil {
		return err
	}

	_, err = manifest.FetchContext()
	if err != nil {
		return err
	}

	p.InitPluralClient()
	return ui.Run(p.Client, c)
}
