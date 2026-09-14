// Package theme owns the semantic terminal palette used by TUI screens. Raw
// Console colors stop here so screens can describe intent instead of styling.
package theme

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

// Colors is the small semantic subset needed by the shell. It is derived from
// Console's dark semantic palette and Cloud Shell accents.
type Colors struct {
	Background color.Color
	Surface    color.Color
	Border     color.Color
	Text       color.Color
	Muted      color.Color
	Primary    color.Color
	Info       color.Color
	Success    color.Color
	Warning    color.Color
	Danger     color.Color
}

// Theme bundles semantic colors and reusable shell styles.
type Theme struct {
	Colors     Colors
	Logo       lipgloss.Style
	Title      lipgloss.Style
	Body       lipgloss.Style
	Muted      lipgloss.Style
	Success    lipgloss.Style
	Warning    lipgloss.Style
	Danger     lipgloss.Style
	Link       lipgloss.Style
	Color      bool
	Hyperlinks bool
}

// New down-samples the Console palette to the terminal's supported profile.
// ASCII is also the explicit NO_COLOR representation.
func New(profile colorprofile.Profile) Theme {
	resolve := func(hex string) color.Color {
		if profile <= colorprofile.ASCII {
			return lipgloss.NoColor{}
		}
		return profile.Convert(lipgloss.Color(hex))
	}

	colors := Colors{
		Background: resolve("#12151B"), // fill-zero
		Surface:    resolve("#1B1F27"), // fill-one
		Border:     resolve("#252932"), // border
		Text:       resolve("#EEF0F1"), // text
		Muted:      resolve("#A1A5B0"), // text-xlight
		Primary:    resolve("#747AF6"), // icon-primary
		Info:       resolve("#99DAFF"), // semanticBlue
		Success:    resolve("#3CECAF"), // cloud-shell-green
		Warning:    resolve("#FFF48F"), // cloud-shell-dark-yellow
		Danger:     resolve("#F2788D"), // cloud-shell-dark-red
	}

	title := lipgloss.NewStyle().Foreground(colors.Primary)
	link := lipgloss.NewStyle().Foreground(colors.Info)
	if profile > colorprofile.ASCII {
		title = title.Bold(true)
		link = link.Underline(true)
	}

	return Theme{
		Colors:     colors,
		Title:      title,
		Logo:       lipgloss.NewStyle().Foreground(colors.Text),
		Body:       lipgloss.NewStyle().Foreground(colors.Text),
		Muted:      lipgloss.NewStyle().Foreground(colors.Muted),
		Success:    lipgloss.NewStyle().Foreground(colors.Success),
		Warning:    lipgloss.NewStyle().Foreground(colors.Warning),
		Danger:     lipgloss.NewStyle().Foreground(colors.Danger),
		Link:       link,
		Color:      profile > colorprofile.ASCII,
		Hyperlinks: profile > colorprofile.ASCII,
	}
}

// Sample renders a stable palette specimen used by golden tests.
func (t Theme) Sample() string {
	return strings.Join([]string{
		t.Title.Render("Plural"),
		t.Body.Render("terminal operations"),
		t.Muted.Render("muted"),
		t.Success.Render("success"),
		t.Warning.Render("warning"),
		t.Danger.Render("danger"),
	}, "\n")
}
