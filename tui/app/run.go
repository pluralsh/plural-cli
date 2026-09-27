package app

import (
	"context"
	"errors"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"

	"github.com/pluralsh/plural-cli/tui/theme"
)

// ErrNoTerminal is returned when the explicit TUI entrypoint has no usable
// interactive terminal.
var ErrNoTerminal = errors.New("plural tui requires an interactive terminal")

// Run validates the terminal contract and starts the root model with
// caller-owned cancellation.
func Run(ctx context.Context, input, output *os.File, dependencies Dependencies) error {
	if input == nil || output == nil || !term.IsTerminal(input.Fd()) || !term.IsTerminal(output.Fd()) {
		return ErrNoTerminal
	}

	profile := colorprofile.Detect(output, os.Environ())
	program := tea.NewProgram(
		New(ctx, theme.New(profile), dependencies),
		tea.WithContext(ctx),
		tea.WithInput(input),
		tea.WithOutput(output),
	)
	_, err := program.Run()
	return err
}
