package theme

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"
)

func TestThemeGoldens(t *testing.T) {
	tests := []struct {
		name    string
		profile colorprofile.Profile
	}{
		{name: "truecolor", profile: colorprofile.TrueColor},
		{name: "ansi256", profile: colorprofile.ANSI256},
		{name: "ansi16", profile: colorprofile.ANSI},
		{name: "no-color", profile: colorprofile.ASCII},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strconv.Quote(New(tt.profile).Sample())
			path := filepath.Join("testdata", tt.name+".golden")
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read golden: %v\nactual: %q", err, got)
			}
			if got != strings.TrimSpace(string(want)) {
				t.Fatalf("theme output differs from %s\nwant: %q\n got: %q", path, want, got)
			}
		})
	}
}
