package tui

func Print(ev Event) {
	events <- ev
}
