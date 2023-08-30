package tui

import (
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

func Start() {
	opts := []tea.ProgramOption{tea.WithoutRenderer()}

	if isatty.IsTerminal(os.Stdout.Fd()) {
		// If we're in TUI mode, discard log output
		opts = []tea.ProgramOption{}
	}

	p := tea.NewProgram(NewTerminalUI(), opts...)
	go func() {
		if _, err := p.StartReturningModel(); err != nil {
			log.Panic(err)
		}
	}()
}

func silence() func() {
	null, _ := os.Open(os.DevNull)
	sout := os.Stdout
	serr := os.Stderr

	color.Output = null
	color.Error = null
	log.SetOutput(io.Discard)

	return func() {
		defer null.Close()
		color.Output = sout
		color.Error = serr
		log.SetOutput(os.Stderr)
	}
}
