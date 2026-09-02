package up

import (
	"context"
	"fmt"
	"io"

	tea "charm.land/bubbletea/v2"

	upbridge "github.com/pluralsh/plural-cli/pkg/bridge/up"
)

// blockingExecCommand runs arbitrary work after tea.Exec releases the terminal
// so CLI/terraform stdout stays line-oriented instead of fighting Bubble Tea redraws.
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

// shouldExecLiveRunner is true for the live runner (writes os.Stdout).
// Test stubs keep an in-process Cmd so Update can receive done messages.
func shouldExecLiveRunner(runner upbridge.Runner) bool {
	if runner == nil {
		return true
	}
	_, ok := runner.(upbridge.LiveRunner)
	return ok
}

// shouldExecDeploy keeps the older name used by beginDeploy.
func shouldExecDeploy(runner upbridge.Runner) bool {
	return shouldExecLiveRunner(runner)
}

func deployExecCmd(ctx context.Context, runner upbridge.Runner, in upbridge.DeployInput) tea.Cmd {
	if runner == nil {
		runner = upbridge.DefaultRunner()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	var steps []string
	var committed string
	in.CommittedMsg = &committed
	cmd := &blockingExecCommand{
		run: func() error {
			fmt.Println()
			fmt.Println("=== Plural up · Deploy (terraform) ===")
			fmt.Println("TUI paused — terraform output uses this terminal.")
			fmt.Println("Commit message is prompted after management terraform (same as plural up).")
			fmt.Println()
			return runner.Deploy(ctx, in, func(step string) {
				steps = append(steps, step)
				fmt.Printf("\n→ %s\n\n", step)
			})
		},
	}
	return tea.Exec(cmd, func(err error) tea.Msg {
		return deployDoneMsg{err: err, steps: steps, commitMsg: committed}
	})
}

func runExecCmd(ctx context.Context, runner upbridge.Runner, in upbridge.RunInput) tea.Cmd {
	if runner == nil {
		runner = upbridge.DefaultRunner()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	var steps []string
	var importID string
	cmd := &blockingExecCommand{
		run: func() error {
			fmt.Println()
			if in.SkipFlush {
				fmt.Println("=== Plural up · Generate ===")
			} else {
				fmt.Println("=== Plural up · Flush + Generate ===")
			}
			fmt.Println("TUI paused — generation output uses this terminal.")
			fmt.Println()
			res, err := runner.Run(ctx, in, func(step string) {
				steps = append(steps, step)
				fmt.Printf("\n→ %s\n\n", step)
			})
			importID = res.ImportClusterID
			return err
		},
	}
	return tea.Exec(cmd, func(err error) tea.Msg {
		return runDoneMsg{err: err, steps: steps, importClusterID: importID}
	})
}
