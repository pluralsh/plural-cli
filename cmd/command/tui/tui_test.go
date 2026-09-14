package tui

import (
	"context"
	"testing"

	"github.com/urfave/cli"
)

func TestCommandIsExplicitLaunchPath(t *testing.T) {
	called := false
	app := cli.NewApp()
	app.Commands = []cli.Command{command(func(context.Context) error {
		called = true
		return nil
	})}

	if err := app.Run([]string{"plural"}); err != nil {
		t.Fatalf("bare app: %v", err)
	}
	if called {
		t.Fatal("bare plural launched the TUI")
	}
	if err := app.Run([]string{"plural", "tui"}); err != nil {
		t.Fatalf("plural tui: %v", err)
	}
	if !called {
		t.Fatal("plural tui did not launch the TUI")
	}
}
