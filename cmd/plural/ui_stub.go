//go:build !ui && !generate

package plural

import (
	"github.com/urfave/cli"
)

func (p *Plural) uiCommands() cli.Command {
	return cli.Command{}
}
