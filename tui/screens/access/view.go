package access

import (
	"fmt"

	"charm.land/lipgloss/v2"

	"github.com/pluralsh/plural-cli/tui/components/page"
)

const profilePanelHeight = 8

func (m Model) View(width, height int) string {
	width, height = page.Size(width, height)
	if width < page.MinimumWidth || height < page.MinimumHeight {
		return page.Unsupported(m.theme, width, height)
	}
	contentWidth := page.ContentWidth(width)
	status := m.headerStatus()

	var body, help string
	switch m.mode {
	case modeDeviceLogin:
		body = page.Panel(m.theme, "Plural App device login", m.deviceLoginLines(), contentWidth, 9, true)
		help = "esc cancel · ctrl+c cancel"
	case modeConsoleForm:
		body = page.Panel(m.theme, "Add Console connection", m.consoleFormLines(), contentWidth, 9, true)
		help = "enter next/save · esc cancel · token is stored securely"
	case modeAccounts:
		body = page.Panel(m.theme, "Choose acting identity", m.accountLines(), contentWidth, 10, true)
		help = "↑/↓ select · enter use for this session · esc cancel"
	default:
		body = m.profileOverview(contentWidth)
		if contentWidth < 100 {
			help = "tab panel · ↑/↓ select · enter use · n App · c Console · i act as · esc back"
		} else {
			help = "tab panel · ↑/↓ select · enter activate · n App login · c Console · i act as · x stop · r refresh · esc back"
		}
	}
	return page.Render(m.theme, width, height, "Identity & connections", status, body, help)
}

func (m Model) headerStatus() string {
	if m.loading {
		return m.theme.Warning.Render("◌ working")
	}
	if m.err != nil {
		return m.theme.Danger.Render("✗ attention required")
	}
	if m.snapshot.Context.Base == nil && m.snapshot.Context.Console == nil {
		return m.theme.Warning.Render("○ setup available")
	}
	return m.theme.Success.Render("✓ context ready")
}

func (m Model) profileOverview(width int) string {
	gap := 1
	leftWidth := (width - gap) / 2
	rightWidth := width - gap - leftWidth
	profiles := page.Panel(m.theme, "Plural App profiles", m.profileLines(), leftWidth, profilePanelHeight, m.panel == 0)
	consoles := page.Panel(m.theme, "Console profiles", m.consoleLines(), rightWidth, profilePanelHeight, m.panel == 1)
	columns := lipgloss.JoinHorizontal(lipgloss.Top, profiles, " ", consoles)
	return columns + "\n\n" + page.Panel(m.theme, "Effective context", m.contextLines(), width, 6, false)
}

func (m Model) profileLines() []string {
	if len(m.snapshot.State.Profiles) == 0 {
		return []string{m.theme.Warning.Render("○ Not connected"), m.theme.Muted.Render("  Press n to sign in with a device code.")}
	}
	lines := make([]string, 0, 2*len(m.snapshot.State.Profiles))
	for i, profile := range m.snapshot.State.Profiles {
		cursor := "  "
		if m.panel == 0 && i == m.appCursor {
			cursor = "› "
		}
		active := ""
		if profile.ID == m.snapshot.State.ActiveProfileID {
			active = "  " + m.theme.Success.Render("ACTIVE")
		}
		lines = append(lines, cursor+profile.Name+active, "    "+m.theme.Muted.Render(profile.Email))
	}
	return lines
}

func (m Model) consoleLines() []string {
	if len(m.snapshot.State.ConsoleProfiles) == 0 {
		return []string{m.theme.Warning.Render("○ Skipped for now"), m.theme.Muted.Render("  Press c to connect later.")}
	}
	lines := make([]string, 0, 2*len(m.snapshot.State.ConsoleProfiles))
	for i, profile := range m.snapshot.State.ConsoleProfiles {
		cursor := "  "
		if m.panel == 1 && i == m.consoleCursor {
			cursor = "› "
		}
		active := ""
		if profile.ID == m.snapshot.State.ActiveConsoleID {
			active = "  " + m.theme.Success.Render("ACTIVE")
		}
		lines = append(lines, cursor+profile.Name+active, "    "+m.theme.Muted.Render(profile.URL))
	}
	return lines
}

func (m Model) contextLines() []string {
	base, acting, console := "not connected", "self", "not connected"
	if m.snapshot.Context.Base != nil {
		base = m.snapshot.Context.Base.Email + " via " + m.snapshot.Context.Base.Name
	}
	if m.snapshot.Context.Acting != nil {
		acting = m.theme.Warning.Render(m.snapshot.Context.Acting.Email) + m.theme.Muted.Render(" · session only")
	}
	if m.snapshot.Context.Console != nil {
		console = m.snapshot.Context.Console.Name + " · " + m.snapshot.Context.Console.URL
	}
	lines := []string{"Base account  " + base, "Acting as     " + acting, "Console       " + console}
	if m.err != nil {
		lines = append(lines, m.theme.Danger.Render("Error         "+m.err.Error()))
	}
	return lines
}

func (m Model) accountLines() []string {
	if len(m.snapshot.ServiceAccounts) == 0 {
		return []string{m.theme.Warning.Render("○ No service accounts available."), m.theme.Muted.Render("  esc returns to profile selection")}
	}
	lines := make([]string, 0, len(m.snapshot.ServiceAccounts)+1)
	for i, account := range m.snapshot.ServiceAccounts {
		cursor := "  "
		if i == m.accountCursor {
			cursor = "› "
		}
		lines = append(lines, cursor+account.Email)
	}
	lines = append(lines, "", m.theme.Muted.Render("The exchanged credential remains in memory only."))
	return lines
}

func (m Model) consoleFormLines() []string {
	labels := []string{"Name", "URL", "Token"}
	lines := []string{m.theme.Muted.Render("Console is optional; press esc to finish setup without it."), ""}
	for i := range m.form {
		marker := "  "
		if i == m.formIndex {
			marker = "› "
		}
		lines = append(lines, fmt.Sprintf("%s%-7s %s", marker, labels[i], m.form[i].View()))
	}
	return lines
}

func (m Model) deviceLoginLines() []string {
	return []string{m.theme.Muted.Render("Open this URL in your browser:"), m.theme.Link.Render(m.authorization.LoginURL), "", m.theme.Warning.Render("◌ Waiting for authorization…"), "", m.theme.Muted.Render("Console can be skipped and connected later from this screen.")}
}
