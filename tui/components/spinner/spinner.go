package spinner

import (
	"time"

	"charm.land/bubbles/v2/spinner"

	"github.com/pluralsh/plural-cli/tui/theme"
)

// Mark follows the logo's top-left bracket, center dot, and bottom-right
// bracket. Every frame occupies one terminal cell, so nearby text stays put.
var Mark = spinner.Spinner{
	Frames: []string{"▛", "●", "▟", "●"},
	FPS:    120 * time.Millisecond,
}

func New(t theme.Theme) spinner.Model {
	return spinner.New(
		spinner.WithSpinner(Mark),
		spinner.WithStyle(t.Title),
	)
}
