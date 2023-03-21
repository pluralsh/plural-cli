//go:build ui || generate

package ui

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	Width  = 1024
	Height = 768
)

// Window struct
type Window struct {
	ctx context.Context
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods.
func (this *Window) startup(ctx context.Context) {
	this.ctx = ctx
}

func (this *Window) width() int {
	return Width
}

func (this *Window) height() int {
	return Height
}

// Close closes the application
func (this *Window) Close() {
	runtime.Quit(this.ctx)
}

// NewWindow creates a new App application struct
func NewWindow() *Window {
	return &Window{}
}
