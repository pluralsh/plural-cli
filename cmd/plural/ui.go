package plural

import (
	"github.com/urfave/cli"

	"github.com/pluralsh/plural/pkg/ui"
)

func (p *Plural) ui(_ *cli.Context) error {
	return ui.Run()
}
