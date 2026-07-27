package spinner

import (
	"reflect"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"

	"github.com/pluralsh/plural-cli/tui/theme"
)

func TestMarkFramesHaveStableCompactWidth(t *testing.T) {
	if len(Mark.Frames) == 0 {
		t.Fatal("spinner has no frames")
	}
	for _, frame := range Mark.Frames {
		if width := lipgloss.Width(frame); width != 1 {
			t.Fatalf("frame %q has width %d, want 1", frame, width)
		}
	}
}

func TestNewUsesPluralMark(t *testing.T) {
	model := New(theme.New(colorprofile.ASCII))
	if got, want := model.Spinner.Frames, Mark.Frames; !reflect.DeepEqual(got, want) {
		t.Fatalf("got frames %q, want %q", got, want)
	}
}
