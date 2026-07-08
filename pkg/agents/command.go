package agents

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Command runs external executables from a configured working directory.
//
// Restorers use this boundary to invoke provider CLIs while tests can replace
// it with narrower fakes.
type Command interface {
	// Run executes command with inherited stdio.
	Run(ctx context.Context, command string, args ...string) error
	// Output executes command and returns trimmed combined output.
	Output(ctx context.Context, command string, args ...string) (string, error)
}

type executable struct {
	// dir is the working directory for executed commands.
	dir string
	// env contains additional KEY=value environment entries.
	env []string
}

func (in *executable) Run(ctx context.Context, command string, args ...string) error {
	if len(command) == 0 {
		return fmt.Errorf("empty command")
	}
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = in.dir
	cmd.Env = append(os.Environ(), in.env...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (in *executable) Output(ctx context.Context, command string, args ...string) (string, error) {
	if len(command) == 0 {
		return "", fmt.Errorf("empty command")
	}
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = in.dir
	cmd.Env = append(os.Environ(), in.env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %s", cmd.String(), strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

// Executable returns a Command backed by os/exec.
func Executable(execDir string, env ...string) Command {
	return &executable{
		dir: execDir,
		env: env,
	}
}
