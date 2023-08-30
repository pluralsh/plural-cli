package tui

import (
	"fmt"

	"github.com/muesli/reflow/indent"
)

func (this *TerminalUI) View() string {
	s := "\n" +
		this.spinner.View() + " Creating a cluster\n\n"

	for _, ev := range this.events {
		if len(ev.Name) == 0 {
			continue
		}

		s += fmt.Sprintf("%s %s\n", ev.Name, ev.Took)
	}

	return indent.String(s, 1)
}
