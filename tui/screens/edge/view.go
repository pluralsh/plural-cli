package edge

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"

	"github.com/pluralsh/plural-cli/tui/components/page"
)

func (m Model) View(width, height int) string {
	width, height = page.Size(width, height)
	if width < page.MinimumWidth || height < page.MinimumHeight {
		return page.Unsupported(m.theme, width, height)
	}
	body, help := m.bodyAndHelp(page.ContentWidth(width), height)
	return page.Render(m.theme, width, height, m.title(), m.status(), body, help)
}

func (m Model) title() string {
	switch m.mode {
	case modeImageForm, modeImageReview, modeImageRunning, modeImageResult:
		return "Edge · image"
	case modeFlashForm, modeFlashDevices, modeFlashDevicePath, modeFlashConfirm, modeFlashRunning, modeFlashResult:
		return "Edge · flash"
	default:
		return "Edge"
	}
}

func (m Model) status() string {
	if m.mode == modeImageRunning {
		return m.theme.Warning.Render("◌ building image")
	}
	if m.mode == modeFlashRunning {
		return m.theme.Warning.Render("◌ flashing")
	}
	if m.mode == modeImageResult || m.mode == modeFlashResult {
		if m.err != nil {
			return m.theme.Danger.Render("✗ failed")
		}
		return m.theme.Success.Render("✓ done")
	}
	if m.needsAuth {
		return m.theme.Warning.Render("○ connect Console")
	}
	return m.theme.Muted.Render("plural edge")
}

func (m Model) bodyAndHelp(width, height int) (string, string) {
	switch m.mode {
	case modeImageForm:
		fields := imageFields()
		field := fields[m.field]
		m.inputs[m.field].SetWidth(max(8, width-8))
		return page.Panel(m.theme, field.label, []string{
			m.theme.Muted.Render(fmt.Sprintf("Same as plural edge image --%s  (%d/%d)", field.key, m.field+1, len(fields))),
			"",
			m.inputs[m.field].View(),
		}, width, 8, true), "enter next · esc back"
	case modeImageReview:
		opts := m.imageOptions()
		lines := []string{
			"Output      " + display(opts.OutputDir),
			"Project     " + display(opts.Project),
			"Model       " + display(opts.Model),
			"Username    " + display(opts.Username),
			"Password    " + secret(opts.Password),
			"User        " + display(opts.User),
			"Wi-Fi       " + display(opts.WifiSSID),
			"Cloud cfg   " + display(opts.CloudConfig),
			"Plural cfg  " + display(opts.PluralConfig),
			"OCI URL     " + display(opts.OCIURL),
			"",
			m.theme.Muted.Render("Builds a Kairos ARM image with Docker (privileged)."),
		}
		return page.Panel(m.theme, "Review image", lines, width, 16, true), "enter build · esc edit"
	case modeImageRunning:
		return m.viewRunning(width, height, "Building", "Running plural edge image…", "Docker / image output streams below (TUI stays open).")
	case modeImageResult:
		return m.viewResult(width, height, "Image complete", "✓ Image saved", "✗ Image build failed", "Output  "+display(m.imageOptions().OutputDir))
	case modeFlashForm:
		m.inputs[0].SetWidth(max(8, width-8))
		lines := []string{
			m.theme.Muted.Render("Same as plural edge flash --image  (1/2)"),
			"",
			m.inputs[0].View(),
		}
		if m.imageSuggested {
			lines = append(lines, "", m.theme.Muted.Render("Found image/kairos.img in the current directory."))
		}
		return page.Panel(m.theme, "Image file", lines, width, 10, true), "enter next · esc back"
	case modeFlashDevices:
		return m.viewFlashDevices(width)
	case modeFlashDevicePath:
		m.inputs[1].SetWidth(max(8, width-8))
		return page.Panel(m.theme, "Storage device", []string{
			m.theme.Muted.Render("Same as plural edge flash --device  (custom path)"),
			"",
			m.inputs[1].View(),
		}, width, 8, true), "enter confirm · esc devices"
	case modeFlashConfirm:
		opts := m.flashOptions()
		device := display(opts.Device)
		if picked, ok := m.selectedFlashDevice(); ok {
			device = picked.Label()
		}
		lines := []string{
			m.theme.Danger.Render("This overwrites the storage device."),
			"",
			"Image   " + display(opts.Image),
			"Device  " + device,
			"",
		}
		if m.flashNeedsRoot() {
			lines = append(lines,
				m.theme.Warning.Render("Needs root to write this device."),
				m.theme.Muted.Render("Flash will try sudo/pkexec. If that fails: sudo plural tui"),
				"",
			)
		}
		lines = append(lines, m.theme.Muted.Render("Same as plural edge flash --image --device."))
		return page.Panel(m.theme, "Confirm flash", lines, width, 12, true), "enter flash · esc edit"
	case modeFlashRunning:
		return m.viewFlashing(width)
	case modeFlashResult:
		return m.viewResult(width, height, "Flash complete", "✓ Image flashed", "✗ Flash failed", "Device  "+display(m.flashOptions().Device))
	default:
		lines := []string{
			m.theme.Muted.Render("Prepare a Raspberry Pi image, then write it to a disk."),
			"",
		}
		items := []struct{ number, title, blurb string }{
			{"1", "image", "build Kairos ARM image (Docker)"},
			{"2", "flash", "write image onto a storage device"},
		}
		for i, item := range items {
			cursor := "  "
			if i == m.cursor {
				cursor = "› "
			}
			row := fmt.Sprintf("%s%s  %-8s %s", cursor, item.number, item.title, item.blurb)
			lines = append(lines, ansi.Truncate(row, width-2, "…"))
		}
		return page.Panel(m.theme, "Edge commands", lines, width, 10, true), "↑/↓ select · enter open · 1-2 shortcut · esc welcome"
	}
}

func (m Model) viewFlashDevices(width int) (string, string) {
	lines := []string{
		m.theme.Muted.Render("USB disks detected from sysfs (whole disk, not a partition)."),
		"",
	}
	if m.deviceErr != nil {
		lines = append(lines, m.theme.Danger.Render(m.deviceErr.Error()), "")
	}
	if len(m.devices) == 0 && m.deviceErr == nil {
		lines = append(lines, m.theme.Muted.Render("No USB disks found. Plug in a stick and press r."), "")
	}
	for i, device := range m.devices {
		label := device.Label()
		if device.NeedsRoot {
			label += "  · needs root"
		}
		lines = append(lines, m.deviceRow(width, i, label))
	}
	lines = append(lines, m.deviceRow(width, len(m.devices), "Enter path…"))
	height := min(14, 6+m.flashDeviceCount())
	if height < 10 {
		height = 10
	}
	return page.Panel(m.theme, "Storage device", lines, width, height, true), "↑/↓ select · enter · 1-9 · r refresh · esc back"
}

func (m Model) deviceRow(width, index int, label string) string {
	cursor := "  "
	if index == m.deviceCursor {
		cursor = "› "
	}
	row := fmt.Sprintf("%s%d  %s", cursor, index+1, label)
	return ansi.Truncate(row, max(1, width-2), "…")
}

func (m Model) viewRunning(width, height int, title, status, hint string) (string, string) {
	panelH, logN := logPanelBudget(height, 4)
	lines := []string{
		m.spin.View() + " " + m.theme.Muted.Render(status),
		"",
		m.theme.Muted.Render(hint),
		"",
	}
	lines = append(lines, m.opLogLines(logN, width)...)
	return page.Panel(m.theme, title, lines, width, panelH, true), "↑/↓ · pgup/pgdn scroll · end follow"
}

func (m Model) viewFlashing(width int) (string, string) {
	opts := m.flashOptions()
	lines := []string{
		m.spin.View() + " " + m.theme.Muted.Render("Writing image onto "+display(opts.Device)),
		"",
		m.flashProgressBar(max(20, width-14)),
		m.theme.Muted.Render(m.flashProgressLabel()),
		"",
	}
	if len(m.opLog) == 0 {
		lines = append(lines, m.theme.Muted.Render("Waiting for flash to start…"))
	} else {
		start := 0
		if len(m.opLog) > 4 {
			start = len(m.opLog) - 4
		}
		for _, line := range m.opLog[start:] {
			lines = append(lines, m.theme.Muted.Render(ansi.Truncate(line, max(1, width-4), "…")))
		}
	}
	return page.Panel(m.theme, "Flashing", lines, width, 12, true), "ctrl+c quit"
}

func (m Model) flashProgressBar(width int) string {
	pct := int(m.flashPercent())
	suffix := fmt.Sprintf("  %3d%%", pct)
	barWidth := width - len(suffix)
	if barWidth < 10 {
		barWidth = 10
	}
	filled := barWidth * pct / 100
	if filled > barWidth {
		filled = barWidth
	}
	bar := m.theme.Success.Render(strings.Repeat("█", filled)) + m.theme.Muted.Render(strings.Repeat("░", barWidth-filled))
	return bar + suffix
}

func (m Model) flashPercent() float64 {
	if m.flashTotal <= 0 {
		return 0
	}
	pct := 100 * float64(m.flashWritten) / float64(m.flashTotal)
	if pct > 100 {
		return 100
	}
	if pct < 0 {
		return 0
	}
	return pct
}

func (m Model) flashProgressLabel() string {
	if m.flashTotal <= 0 && m.flashWritten <= 0 {
		return "0 B / —"
	}
	label := formatFlashSize(m.flashWritten) + " / "
	if m.flashTotal > 0 {
		label += formatFlashSize(m.flashTotal)
	} else {
		label += "—"
	}
	if rate := m.flashRate(); rate != "" {
		label += "   " + rate
	}
	if m.flashTotal > 0 {
		label += fmt.Sprintf("   %d%%", int(m.flashPercent()))
	}
	return label
}

func (m Model) flashRate() string {
	if m.flashWritten <= 0 || m.flashStarted.IsZero() {
		return ""
	}
	elapsed := time.Since(m.flashStarted).Seconds()
	if elapsed < 0.2 {
		return ""
	}
	return formatFlashSize(int64(float64(m.flashWritten)/elapsed)) + "/s"
}

func formatFlashSize(n int64) string {
	if n < 0 {
		n = 0
	}
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

func (m Model) viewResult(width, height int, title, ok, fail, detail string) (string, string) {
	panelH, logN := logPanelBudget(height, 6)
	lines := []string{}
	if m.err != nil {
		lines = append(lines,
			m.theme.Danger.Render(fail+" — scroll logs below, then enter/esc hub"),
			m.theme.Danger.Render(m.err.Error()),
			"",
		)
		if m.needsAuth {
			lines = append(lines, m.theme.Muted.Render("Press c to open Access, or retry with a cloud-config path."), "")
		}
	} else {
		lines = append(lines,
			m.theme.Success.Render(ok),
			m.theme.Muted.Render("Scroll logs below · enter/esc hub"),
			detail,
			"",
		)
	}
	lines = append(lines, m.opLogExportHint(width))
	lines = append(lines, m.opLogScrollHint(width))
	lines = append(lines, m.opLogLines(logN, width)...)
	help := "↑/↓ scroll · e export · enter/esc hub"
	if m.needsAuth {
		help = "c connect · ↑/↓ scroll · e export · enter/esc hub"
	}
	return page.Panel(m.theme, title, lines, width, panelH, true), help
}

func display(v string) string {
	if strings.TrimSpace(v) == "" {
		return "—"
	}
	return v
}

func secret(v string) string {
	if strings.TrimSpace(v) == "" {
		return "—"
	}
	return "••••"
}

func (m Model) opLogLines(limit, width int) []string {
	if limit <= 0 {
		limit = 12
	}
	if len(m.opLog) == 0 {
		return []string{m.theme.Muted.Render("Waiting for output…")}
	}
	wrapped := wrapOpLog(m.opLog, width)
	start := m.opLogStart(limit, len(wrapped))
	end := start + limit
	if end > len(wrapped) {
		end = len(wrapped)
	}
	out := make([]string, 0, end-start)
	for _, line := range wrapped[start:end] {
		out = append(out, m.theme.Muted.Render(line))
	}
	return out
}

func opLogInnerWidth(panelWidth int) int {
	return max(20, panelWidth-4)
}

func wrapOpLog(lines []string, width int) []string {
	maxW := opLogInnerWidth(width)
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			out = append(out, "")
			continue
		}
		out = append(out, strings.Split(ansi.Wrap(line, maxW, ""), "\n")...)
	}
	return out
}

func opLogContentWidth(termWidth int) int {
	if termWidth <= 0 {
		termWidth = page.DefaultWidth
	}
	return page.ContentWidth(termWidth)
}

func (m Model) opLogStart(limit, total int) int {
	if total == 0 {
		return 0
	}
	maxStart := max(0, total-limit)
	if m.opLogFollow {
		return maxStart
	}
	if m.opLogY < 0 {
		return 0
	}
	if m.opLogY > maxStart {
		return maxStart
	}
	return m.opLogY
}

func (m Model) opLogScrollHint(width int) string {
	if len(m.opLog) == 0 {
		return m.theme.Muted.Render("No log lines captured.")
	}
	limit := 12
	if m.viewH > 0 {
		_, limit = logPanelBudget(m.viewH, 5)
	}
	wrapped := wrapOpLog(m.opLog, width)
	start := m.opLogStart(limit, len(wrapped))
	end := start + limit
	if end > len(wrapped) {
		end = len(wrapped)
	}
	label := fmt.Sprintf("Logs  %d–%d / %d", start+1, end, len(wrapped))
	if m.opLogFollow {
		label += "  · following"
	}
	return m.theme.Muted.Render(ansi.Truncate(label, max(1, width-4), "…"))
}

func (m Model) opLogExportHint(width int) string {
	if m.logExportErr != nil {
		return m.theme.Danger.Render(ansi.Truncate("Could not save logs: "+m.logExportErr.Error()+" · e retry", max(1, width-4), "…"))
	}
	if m.logExportPath != "" {
		return m.theme.Muted.Render(ansi.Truncate("Saved "+m.logExportPath+" · e to save again", max(1, width-4), "…"))
	}
	return m.theme.Muted.Render("e exports full logs to a file (ctrl+c quits the TUI)")
}

func logPanelBudget(termHeight, chrome int) (panelHeight, logLines int) {
	panelHeight = termHeight - 6
	if panelHeight < 18 {
		panelHeight = 18
	}
	if chrome < 0 {
		chrome = 0
	}
	logLines = panelHeight - 2 - chrome
	if logLines < 12 {
		logLines = 12
	}
	return panelHeight, logLines
}
