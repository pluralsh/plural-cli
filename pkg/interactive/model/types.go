package model

import (
	"context"

	"github.com/derailed/tview"
)

// Igniter represents a runnable view.
type Igniter interface {
	// Start starts a component.
	Init(ctx context.Context) error

	// Start starts a component.
	Start()

	// Stop terminates a component.
	Stop()
}

// Hinter represent a menu mnemonic provider.
type Hinter interface {
	// Hints returns a collection of menu hints.
	Hints() MenuHints

	// ExtraHints returns additional hints.
	ExtraHints() map[string]string
}

// Primitive represents a UI primitive.
type Primitive interface {
	tview.Primitive

	// Name returns the view name.
	Name() string
}

// Commander tracks prompt status.
type Commander interface {
	// InCmdMode checks if prompt is active.
	InCmdMode() bool
}

type Component interface {
	Primitive
	Igniter
	Hinter
	Commander
}
