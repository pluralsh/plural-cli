package config

import (
	"fmt"
	"os"

	"github.com/derailed/tview"
	"github.com/gdamore/tcell/v2"
	"gopkg.in/yaml.v2"
)

// K9sStylesFile represents K9s skins file location.
// var K9sStylesFile = filepath.Join(K9sHome(), "skin.yml")

// StyleListener represents a skin's listener.
type StyleListener interface {
	// StylesChanged notifies listener the skin changed.
	StylesChanged(*Styles)
}

type (
	// Color represents a color.
	Color string

	// Colors tracks multiple colors.
	Colors []Color

	// Styles tracks K9s styling options.
	Styles struct {
		K9s       Style
		listeners []StyleListener
	}

	// Style tracks K9s styles.
	Style struct {
		Body   Body
		Prompt Prompt
		Help   Help
		Frame  Frame
		Info   Info
		Views  Views
		Dialog Dialog
	}

	// Prompt tracks command styles
	Prompt struct {
		FgColor      Color
		BgColor      Color
		SuggestColor Color
	}

	// Help tracks help styles.
	Help struct {
		FgColor      Color
		BgColor      Color
		SectionColor Color
		KeyColor     Color
		NumKeyColor  Color
	}

	// Body tracks body styles.
	Body struct {
		FgColor   Color
		BgColor   Color
		LogoColor Color
	}

	// Dialog tracks dialog styles.
	Dialog struct {
		FgColor            Color
		BgColor            Color
		ButtonFgColor      Color
		ButtonBgColor      Color
		ButtonFocusFgColor Color
		ButtonFocusBgColor Color
		LabelFgColor       Color
		FieldFgColor       Color
	}

	// Frame tracks frame styles.
	Frame struct {
		Title  Title
		Border Border
		Menu   Menu
		Crumb  Crumb
		Status Status
	}

	// Views tracks individual view styles.
	Views struct {
		Table  Table
		Xray   Xray
		Charts Charts
		Yaml   Yaml
		Log    Log
	}

	// Status tracks resource status styles.
	Status struct {
		NewColor       Color
		ModifyColor    Color
		AddColor       Color
		PendingColor   Color
		ErrorColor     Color
		HighlightColor Color
		KillColor      Color
		CompletedColor Color
	}

	// Log tracks Log styles.
	Log struct {
		FgColor   Color
		BgColor   Color
		Indicator LogIndicator
	}

	// LogIndicator tracks log view indicator.
	LogIndicator struct {
		FgColor Color
		BgColor Color
	}

	// Yaml tracks yaml styles.
	Yaml struct {
		KeyColor   Color
		ValueColor Color
		ColonColor Color
	}

	// Title tracks title styles.
	Title struct {
		FgColor        Color
		BgColor        Color
		HighlightColor Color
		CounterColor   Color
		FilterColor    Color
	}

	// Info tracks info styles.
	Info struct {
		SectionColor Color
		FgColor      Color
	}

	// Border tracks border styles.
	Border struct {
		FgColor    Color
		FocusColor Color
	}

	// Crumb tracks crumbs styles.
	Crumb struct {
		FgColor     Color
		BgColor     Color
		ActiveColor Color
	}

	// Table tracks table styles.
	Table struct {
		FgColor       Color
		BgColor       Color
		CursorFgColor Color
		CursorBgColor Color
		MarkColor     Color
		Header        TableHeader
	}

	// TableHeader tracks table header styles.
	TableHeader struct {
		FgColor     Color
		BgColor     Color
		SorterColor Color
	}

	// Xray tracks xray styles.
	Xray struct {
		FgColor         Color
		BgColor         Color
		CursorColor     Color
		CursorTextColor Color
		GraphicColor    Color
	}

	// Menu tracks menu styles.
	Menu struct {
		FgColor     Color
		KeyColor    Color
		NumKeyColor Color
	}

	// Charts tracks charts styles.
	Charts struct {
		BgColor            Color
		DialBgColor        Color
		ChartBgColor       Color
		DefaultDialColors  Colors
		DefaultChartColors Colors
		ResourceColors     map[string]Colors
	}
)

const (
	// DefaultColor represents  a default color.
	DefaultColor Color = "default"

	// TransparentColor represents the terminal bg color.
	TransparentColor Color = "-"
)

// NewColor returns a new color.
func NewColor(c string) Color {
	return Color(c)
}

// String returns color as string.
func (c Color) String() string {
	if c.isHex() {
		return string(c)
	}
	if c == DefaultColor {
		return "-"
	}
	col := c.Color().TrueColor().Hex()
	if col < 0 {
		return "-"
	}

	return fmt.Sprintf("#%06x", col)
}

func (c Color) isHex() bool {
	return len(c) == 7 && c[0] == '#'
}

// Color returns a view color.
func (c Color) Color() tcell.Color {
	if c == DefaultColor {
		return tcell.ColorDefault
	}

	return tcell.GetColor(string(c)).TrueColor()
}

// Colors converts series string colors to colors.
func (c Colors) Colors() []tcell.Color {
	cc := make([]tcell.Color, 0, len(c))
	for _, color := range c {
		cc = append(cc, color.Color())
	}
	return cc
}

func newStyle() Style {
	return Style{
		Body:   newBody(),
		Prompt: newPrompt(),
		Help:   newHelp(),
		Frame:  newFrame(),
		Info:   newInfo(),
		Views:  newViews(),
		Dialog: newDialog(),
	}
}

func newDialog() Dialog {
	return Dialog{
		FgColor:            "cadetBlue",
		BgColor:            "black",
		ButtonBgColor:      "darkslateblue",
		ButtonFgColor:      "black",
		ButtonFocusBgColor: "dodgerblue",
		ButtonFocusFgColor: "black",
		LabelFgColor:       "white",
		FieldFgColor:       "white",
	}
}

func newPrompt() Prompt {
	return Prompt{
		FgColor:      "cadetBlue",
		BgColor:      "black",
		SuggestColor: "dodgerblue",
	}
}

func newCharts() Charts {
	return Charts{
		BgColor:            "black",
		DialBgColor:        "black",
		ChartBgColor:       "black",
		DefaultDialColors:  Colors{Color("palegreen"), Color("orangered")},
		DefaultChartColors: Colors{Color("palegreen"), Color("orangered")},
		ResourceColors: map[string]Colors{
			"cpu": {Color("dodgerblue"), Color("darkslateblue")},
			"mem": {Color("yellow"), Color("goldenrod")},
		},
	}
}

func newViews() Views {
	return Views{
		Table:  newTable(),
		Xray:   newXray(),
		Charts: newCharts(),
		Yaml:   newYaml(),
		Log:    newLog(),
	}
}

func newFrame() Frame {
	return Frame{
		Title:  newTitle(),
		Border: newBorder(),
		Menu:   newMenu(),
		Crumb:  newCrumb(),
		Status: newStatus(),
	}
}

func newHelp() Help {
	return Help{
		FgColor:      "cadetblue",
		BgColor:      "black",
		SectionColor: "green",
		KeyColor:     "dodgerblue",
		NumKeyColor:  "fuchsia",
	}
}

func newBody() Body {
	return Body{
		FgColor:   "cadetblue",
		BgColor:   "black",
		LogoColor: "orange",
	}
}

func newStatus() Status {
	return Status{
		NewColor:       "lightskyblue",
		ModifyColor:    "greenyellow",
		AddColor:       "dodgerblue",
		PendingColor:   "darkorange",
		ErrorColor:     "orangered",
		HighlightColor: "aqua",
		KillColor:      "mediumpurple",
		CompletedColor: "lightslategray",
	}
}

func newLog() Log {
	return Log{
		FgColor:   "lightskyblue",
		BgColor:   "black",
		Indicator: newLogIndicator(),
	}
}

func newLogIndicator() LogIndicator {
	return LogIndicator{
		FgColor: "dodgerblue",
		BgColor: "black",
	}
}

func newYaml() Yaml {
	return Yaml{
		KeyColor:   "steelblue",
		ColonColor: "white",
		ValueColor: "papayawhip",
	}
}

func newTitle() Title {
	return Title{
		FgColor:        "aqua",
		BgColor:        "black",
		HighlightColor: "fuchsia",
		CounterColor:   "papayawhip",
		FilterColor:    "seagreen",
	}
}

func newInfo() Info {
	return Info{
		SectionColor: "white",
		FgColor:      "orange",
	}
}

func newXray() Xray {
	return Xray{
		FgColor:         "aqua",
		BgColor:         "black",
		CursorColor:     "dodgerblue",
		CursorTextColor: "black",
		GraphicColor:    "cadetblue",
	}
}

func newTable() Table {
	return Table{
		FgColor:       "aqua",
		BgColor:       "black",
		CursorFgColor: "black",
		CursorBgColor: "aqua",
		MarkColor:     "palegreen",
		Header:        newTableHeader(),
	}
}

func newTableHeader() TableHeader {
	return TableHeader{
		FgColor:     "white",
		BgColor:     "black",
		SorterColor: "aqua",
	}
}

func newCrumb() Crumb {
	return Crumb{
		FgColor:     "black",
		BgColor:     "aqua",
		ActiveColor: "orange",
	}
}

func newBorder() Border {
	return Border{
		FgColor:    "dodgerblue",
		FocusColor: "lightskyblue",
	}
}

func newMenu() Menu {
	return Menu{
		FgColor:     "white",
		KeyColor:    "dodgerblue",
		NumKeyColor: "fuchsia",
	}
}

// NewStyles creates a new default config.
func NewStyles() *Styles {
	return &Styles{
		K9s: newStyle(),
	}
}

// Reset resets styles.
func (s *Styles) Reset() {
	s.K9s = newStyle()
}

// DefaultSkin loads the default skin.
func (s *Styles) DefaultSkin() {
	s.K9s = newStyle()
}

// FgColor returns the foreground color.
func (s *Styles) FgColor() tcell.Color {
	return s.Body().FgColor.Color()
}

// BgColor returns the background color.
func (s *Styles) BgColor() tcell.Color {
	return s.Body().BgColor.Color()
}

// AddListener registers a new listener.
func (s *Styles) AddListener(l StyleListener) {
	s.listeners = append(s.listeners, l)
}

// RemoveListener removes a listener.
func (s *Styles) RemoveListener(l StyleListener) {
	victim := -1
	for i, lis := range s.listeners {
		if lis == l {
			victim = i
			break
		}
	}
	if victim == -1 {
		return
	}
	s.listeners = append(s.listeners[:victim], s.listeners[victim+1:]...)
}

func (s *Styles) fireStylesChanged() {
	for _, list := range s.listeners {
		list.StylesChanged(s)
	}
}

// Body returns body styles.
func (s *Styles) Body() Body {
	return s.K9s.Body
}

// Frame returns frame styles.
func (s *Styles) Frame() Frame {
	return s.K9s.Frame
}

// Crumb returns crumb styles.
func (s *Styles) Crumb() Crumb {
	return s.Frame().Crumb
}

// Title returns title styles.
func (s *Styles) Title() Title {
	return s.Frame().Title
}

// Charts returns charts styles.
func (s *Styles) Charts() Charts {
	return s.K9s.Views.Charts
}

// Dialog returns dialog styles.
func (s *Styles) Dialog() Dialog {
	return s.K9s.Dialog
}

// Table returns table styles.
func (s *Styles) Table() Table {
	return s.K9s.Views.Table
}

// Xray returns xray styles.
func (s *Styles) Xray() Xray {
	return s.K9s.Views.Xray
}

// Views returns views styles.
func (s *Styles) Views() Views {
	return s.K9s.Views
}

// Load K9s configuration from file.
func (s *Styles) Load(path string) error {
	f, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(f, s); err != nil {
		return err
	}
	// s.fireStylesChanged()

	return nil
}

// Update apply terminal colors based on styles.
func (s *Styles) Update() {
	tview.Styles.PrimitiveBackgroundColor = s.BgColor()
	tview.Styles.ContrastBackgroundColor = s.BgColor()
	tview.Styles.MoreContrastBackgroundColor = s.BgColor()
	tview.Styles.PrimaryTextColor = s.FgColor()
	tview.Styles.BorderColor = s.K9s.Frame.Border.FgColor.Color()
	tview.Styles.FocusColor = s.K9s.Frame.Border.FocusColor.Color()
	tview.Styles.TitleColor = s.FgColor()
	tview.Styles.GraphicsColor = s.FgColor()
	tview.Styles.SecondaryTextColor = s.FgColor()
	tview.Styles.TertiaryTextColor = s.FgColor()
	tview.Styles.InverseTextColor = s.FgColor()
	tview.Styles.ContrastSecondaryTextColor = s.FgColor()

	s.fireStylesChanged()
}
