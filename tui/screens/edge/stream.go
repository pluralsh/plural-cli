package edge

import tea "charm.land/bubbletea/v2"

const maxOpLogLines = 500

type opLogLineMsg struct{ line string }

type flashProgressMsg struct {
	written, total int64
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

func listenFlashProgress(ch <-chan flashProgressMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

func appendOpLog(lines []string, line string) []string {
	lines = append(lines, line)
	if len(lines) > maxOpLogLines {
		lines = lines[len(lines)-maxOpLogLines:]
	}
	return lines
}

func sendLog(ch chan<- string, line string) {
	if ch == nil || line == "" {
		return
	}
	select {
	case ch <- line:
	default:
	}
}

func sendFlashProgress(ch chan flashProgressMsg, msg flashProgressMsg) {
	if ch == nil {
		return
	}
	select {
	case ch <- msg:
	default:
		select {
		case <-ch:
		default:
		}
		select {
		case ch <- msg:
		default:
		}
	}
}
