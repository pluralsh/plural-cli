package down

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

func shouldExecLiveRunner(runner upbridge.Runner) bool {
	if runner == nil {
		return true
	}
	_, ok := runner.(upbridge.LiveRunner)
	return ok
}

func appendOpLog(lines []string, line string) []string {
	lines = append(lines, line)
	if len(lines) > maxOpLogLines {
		lines = lines[len(lines)-maxOpLogLines:]
	}
	return lines
}

func destroyStreamCmd(ctx context.Context, runner upbridge.Runner, in upbridge.DestroyInput, lines chan string) tea.Cmd {
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
		err := runner.Destroy(ctx, in, func(step string) {
			steps = append(steps, step)
			_, _ = io.WriteString(w, "→ "+step+"\n")
		})
		w.Close()
		close(lines)
		return destroyDoneMsg{err: err, steps: steps}
	}
}
