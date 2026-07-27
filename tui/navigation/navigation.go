// Package navigation defines route messages shared by otherwise independent
// screens. Keeping route ownership out of individual screens lets new features
// be developed without importing the root application package.
package navigation

import tea "charm.land/bubbletea/v2"

// Route identifies a top-level TUI screen.
type Route string

const (
	Welcome     Route = "welcome"
	Access      Route = "access"
	Diagnostics Route = "diagnostics"
)

// NavigateMsg requests a top-level route change.
type NavigateMsg struct{ Route Route }

// Navigate returns a typed route command.
func Navigate(route Route) tea.Cmd { return func() tea.Msg { return NavigateMsg{Route: route} } }
