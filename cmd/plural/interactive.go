package main

import (
	// ui "github.com/pluralsh/plural/pkg/ui/old"
	"github.com/pluralsh/plural/pkg/interactive/view"
	"github.com/urfave/cli"
)

func handleInteractive(c *cli.Context) error {
	// return ui.InteractiveLayout(c)
	app := view.NewApp(c)
	return app.Run()
}
