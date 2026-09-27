package up

import (
	"fmt"
	"io"

	tea "charm.land/bubbletea/v2"
)

// scmExecCommand runs SCM device login / create / clone after the TUI releases
// the terminal (tea.Exec → releaseTerminal), matching plural up's scm.Setup.
type scmExecCommand struct {
	providerID string
	setup      func(string) (string, error)
	repoName   string
}

func (c *scmExecCommand) Run() error {
	name, err := c.setup(c.providerID)
	c.repoName = name
	return err
}

func (c *scmExecCommand) SetStdin(io.Reader)  {}
func (c *scmExecCommand) SetStdout(io.Writer) {}
func (c *scmExecCommand) SetStderr(io.Writer) {}

// scmSetupCmd returns either a normal Cmd (tests) or tea.Exec (live oauth/survey).
func scmSetupCmd(providerID string, setup func(string) (string, error), useExec bool) tea.Cmd {
	if setup == nil {
		return func() tea.Msg {
			return scmDoneMsg{err: fmt.Errorf("scm setup is not configured")}
		}
	}
	if !useExec {
		return func() tea.Msg {
			name, err := setup(providerID)
			return scmDoneMsg{repo: name, err: err}
		}
	}
	cmd := &scmExecCommand{providerID: providerID, setup: setup}
	return tea.Exec(cmd, func(err error) tea.Msg {
		return scmDoneMsg{repo: cmd.repoName, err: err}
	})
}
