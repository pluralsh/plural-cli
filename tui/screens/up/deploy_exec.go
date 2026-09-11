package up

import (
	"context"
	"io"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"

	upbridge "github.com/pluralsh/plural-cli/pkg/bridge/up"
)

const maxOpLogLines = 500

type opLogLineMsg struct{ line string }

type commitNeededMsg struct{}

// commitGate blocks Deploy at the commit checkpoint until the TUI replies.
type commitGate struct {
	req   chan struct{}
	reply chan string
}

func newCommitGate() *commitGate {
	return &commitGate{
		req:   make(chan struct{}),
		reply: make(chan string),
	}
}

func (g *commitGate) Prompt() string {
	g.req <- struct{}{}
	return <-g.reply
}

// lineWriter splits writes into lines and pushes them onto ch (non-blocking when full).
type lineWriter struct {
	ch  chan<- string
	mu  sync.Mutex
	buf strings.Builder
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, b := range p {
		if b == '\n' {
			w.flushLocked()
			continue
		}
		if b == '\r' {
			continue
		}
		w.buf.WriteByte(b)
	}
	return len(p), nil
}

func (w *lineWriter) flushLocked() {
	line := strings.TrimRight(w.buf.String(), "\r")
	w.buf.Reset()
	if line == "" {
		return
	}
	select {
	case w.ch <- line:
	default:
		// drop if UI is behind — prefer keeping terraform unblocked
	}
}

func (w *lineWriter) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.buf.Len() > 0 {
		w.flushLocked()
	}
}

func listenOpLog(ch <-chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return nil
		}
		return opLogLineMsg{line: line}
	}
}

func listenCommitRequest(g *commitGate) tea.Cmd {
	if g == nil {
		return nil
	}
	return func() tea.Msg {
		_, ok := <-g.req
		if !ok {
			return nil
		}
		return commitNeededMsg{}
	}
}

// shouldStreamLiveRunner prefers in-TUI log capture for the live runner.
// Test stubs keep a simple in-process Cmd (no log channel).
func shouldStreamLiveRunner(runner upbridge.Runner) bool {
	if runner == nil {
		return true
	}
	_, ok := runner.(upbridge.LiveRunner)
	return ok
}

func shouldExecDeploy(runner upbridge.Runner) bool {
	return shouldStreamLiveRunner(runner)
}

func appendOpLog(lines []string, line string) []string {
	lines = append(lines, line)
	if len(lines) > maxOpLogLines {
		lines = lines[len(lines)-maxOpLogLines:]
	}
	return lines
}

func deployStreamCmd(ctx context.Context, runner upbridge.Runner, in upbridge.DeployInput, lines chan string, gate *commitGate) tea.Cmd {
	if runner == nil {
		runner = upbridge.DefaultRunner()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return func() tea.Msg {
		w := &lineWriter{ch: lines}
		in.Output = w
		var steps []string
		var committed string
		in.CommittedMsg = &committed
		if gate != nil {
			in.PromptCommit = true
			in.CommitPrompt = gate.Prompt
		}
		err := runner.Deploy(ctx, in, func(step string) {
			steps = append(steps, step)
			_, _ = io.WriteString(w, "→ "+step+"\n")
		})
		w.Close()
		close(lines)
		if gate != nil {
			close(gate.req)
		}
		return deployDoneMsg{err: err, steps: steps, commitMsg: committed}
	}
}

func runStreamCmd(ctx context.Context, runner upbridge.Runner, in upbridge.RunInput, lines chan string) tea.Cmd {
	if runner == nil {
		runner = upbridge.DefaultRunner()
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return func() tea.Msg {
		w := &lineWriter{ch: lines}
		in.Output = w
		var steps []string
		var importID string
		res, err := runner.Run(ctx, in, func(step string) {
			steps = append(steps, step)
			_, _ = io.WriteString(w, "→ "+step+"\n")
		})
		importID = res.ImportClusterID
		w.Close()
		close(lines)
		return runDoneMsg{err: err, steps: steps, importClusterID: importID}
	}
}
