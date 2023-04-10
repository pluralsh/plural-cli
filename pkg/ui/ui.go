//go:build ui || generate

package ui

import (
	"embed"

	"github.com/urfave/cli"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/pluralsh/plural/pkg/api"
)

//go:embed all:web/dist
var assets embed.FS

func Run(c api.Client, ctx *cli.Context) error {
	// Create an instance of the main window structure
	window := NewWindow()
	client := NewClient(c, ctx)

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "Plural",
		Frameless: true,
		Width:     window.width(),
		Height:    window.height(),
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: window.startup,
		Bind: []interface{}{
			window,
			client,
		},
	})

	return err
}
