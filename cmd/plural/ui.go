//go:build ui

package plural

import (
	"github.com/urfave/cli"

	"github.com/pluralsh/plural/pkg/ui"
)

func UICLICommand() cli.Command {
	return cli.Command{
		Name:   "ui",
		Usage:  "todo",
		Action: run,
	}
}

func run(_ *cli.Context) error {
	return ui.Run()
}
