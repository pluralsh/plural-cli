package main

import (
	"github.com/pluralsh/plural/pkg/ui"
	"github.com/urfave/cli"
)

func handleInteractive(c *cli.Context) error {
	return ui.InteractiveLayout(c)
}
