// Package oplog writes captured TUI terraform/generation logs to a file so
// they can be copied after the terminal UI exits.
package oplog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
)

// Write saves kind logs (and optional error) to dir. Empty dir uses the
// working directory, then the system temp dir if that write fails.
func Write(dir, kind string, lines []string, runErr error) (string, error) {
	if dir == "" {
		wd, err := os.Getwd()
		if err != nil {
			dir = os.TempDir()
		} else {
			dir = wd
		}
	}
	name := fmt.Sprintf("plural-%s-%s.log", kind, time.Now().Format("20060102-150405"))
	body := format(kind, lines, runErr)
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		if dir != os.TempDir() {
			alt := filepath.Join(os.TempDir(), name)
			if err2 := os.WriteFile(alt, []byte(body), 0o644); err2 == nil {
				return alt, nil
			}
		}
		return "", err
	}
	return path, nil
}

func format(kind string, lines []string, runErr error) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# plural %s log  %s\n", kind, time.Now().Format(time.RFC3339))
	if runErr != nil {
		fmt.Fprintf(&b, "# error: %s\n", runErr.Error())
	}
	b.WriteByte('\n')
	for _, line := range lines {
		b.WriteString(ansi.Strip(line))
		b.WriteByte('\n')
	}
	return b.String()
}
