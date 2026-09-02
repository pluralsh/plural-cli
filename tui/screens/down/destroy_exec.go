package down

import (
	"context"
	"fmt"
	"io"

	tea "charm.land/bubbletea/v2"

	upbridge "github.com/pluralsh/plural-cli/pkg/bridge/up"
)

// blockingExecCommand runs work after tea.Exec releases the terminal so
// terraform stdout stays line-oriented instead of fighting Bubble Tea redraws.
type blockingExecCommand struct {
	run func() error
}

func (c *blockingExecCommand) Run() error {
	if c.run == nil {
		return fmt.Errorf("exec run is not configured")
	}
	return c.run()
}

func (c *blockingExecCommand) SetStdin(io.Reader)  {}
func (c *blockingExecCommand) SetStdout(io.Writer) {}
func (c *blockingExecCommand) SetStderr(io.Writer) {}

func shouldExecLiveRunner(runner upbridge.Runner) bool {
	if runner == nil {
		return true
	}
	_, ok := runner.(upbridge.LiveRunner)
	return ok
}

func destroyExecCmd(ctx context.Context, runner upbridge.Runner, in upbridge.DestroyInput) tea.Cmd {
	if runner == nil {
		runner = upbridge.DefaultRunner()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	var steps []string
	cmd := &blockingExecCommand{
		run: func() error {
			fmt.Println()
			fmt.Println("=== Plural down · Destroy (terraform) ===")
			fmt.Println("TUI paused — terraform output uses this terminal.")
			fmt.Println()
			return runner.Destroy(ctx, in, func(step string) {
				steps = append(steps, step)
				fmt.Printf("\n→ %s\n\n", step)
			})
		},
	}
	return tea.Exec(cmd, func(err error) tea.Msg {
		return destroyDoneMsg{err: err, steps: steps}
	})
}
