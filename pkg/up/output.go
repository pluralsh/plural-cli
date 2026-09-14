package up

import (
	"io"
	"os"
	"sync"
)

var (
	commandOutputMu sync.RWMutex
	commandStdout   io.Writer = os.Stdout
	commandStderr   io.Writer = os.Stderr
)

// SetCommandOutput redirects terraform (and related) command stdout/stderr.
// Pass nil writers to restore os.Stdout / os.Stderr. Safe for concurrent use.
func SetCommandOutput(stdout, stderr io.Writer) {
	commandOutputMu.Lock()
	defer commandOutputMu.Unlock()
	if stdout == nil {
		commandStdout = os.Stdout
	} else {
		commandStdout = stdout
	}
	if stderr == nil {
		commandStderr = os.Stderr
	} else {
		commandStderr = stderr
	}
}

func commandOutput() (stdout, stderr io.Writer) {
	commandOutputMu.RLock()
	defer commandOutputMu.RUnlock()
	return commandStdout, commandStderr
}

// commandStdoutWriter is a convenience for fmt.Fprint-style generation messages.
func commandStdoutWriter() io.Writer {
	stdout, _ := commandOutput()
	return stdout
}
