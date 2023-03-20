//go:build !ui

package plural

import (
	"github.com/urfave/cli"
)

func UICLICommand() cli.Command {
	return cli.Command{}
}
