package agents

import (
	"strings"
	"testing"
)

func TestExecutableRunIncludesStderr(t *testing.T) {
	err := Executable(".").Run(t.Context(), "sh", "-c", "echo boom >&2; exit 7")
	if err == nil {
		t.Fatal("expected command failure")
	}
	got := err.Error()
	if !strings.Contains(got, "boom") || !strings.Contains(got, "exit status 7") {
		t.Fatalf("error %q missing stderr or exit status", got)
	}
}

func TestExecutableRunWrapsEmptyStderr(t *testing.T) {
	err := Executable(".").Run(t.Context(), "false")
	if err == nil {
		t.Fatal("expected command failure")
	}
	if !strings.Contains(err.Error(), "false") {
		t.Fatalf("error %q missing command", err)
	}
}
