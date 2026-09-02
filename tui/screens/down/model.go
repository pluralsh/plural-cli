// Package down implements the plural-down destroy wizard.
package down

import (
	"context"
	"fmt"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	upbridge "github.com/pluralsh/plural-cli/pkg/bridge/up"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/utils"
	pluralspinner "github.com/pluralsh/plural-cli/tui/components/spinner"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type mode uint8

const (
	modeSelectCloud mode = iota
	modeAffirm
	modeDestroying
	modeComplete
)

type keyAction uint8

const (
	keyActionNone keyAction = iota
	keyActionUp
	keyActionDown
	keyActionConfirm
	keyActionBack
)

var keyActionKeystrokes = map[keyAction]string{
	keyActionUp:      "up",
	keyActionDown:    "down",
	keyActionConfirm: "enter",
	keyActionBack:    "esc",
}

func actionForKeystroke(keystroke string) keyAction {
	for action, candidate := range keyActionKeystrokes {
		if keystroke == candidate {
			return action
		}
	}
	return keyActionNone
}

type cloudOption struct {
	cloud bool
	id    string
	title string
	blurb string
	cli   string
}

func cloudOptions() []cloudOption {
	return []cloudOption{
		{cloud: false, id: "self-hosted", title: "Self-hosted", blurb: "destroy terraform/mgmt (default)", cli: "plural down"},
		{cloud: true, id: "cloud", title: "Plural Cloud", blurb: "state-rm plural_cluster.mgmt then destroy (--cloud)", cli: "plural down --cloud"},
	}
}

func affirmOptions() []struct {
	value bool
	title string
	blurb string
} {
	return []struct {
		value bool
		title string
		blurb string
	}{
		{value: true, title: "Yes", blurb: "destroy management cluster (default)"},
		{value: false, title: "No", blurb: "cancel — leave infrastructure intact"},
	}
}

type destroyDoneMsg struct {
	err   error
	steps []string
}

// Model is the Down wizard screen.
type Model struct {
	ctx    context.Context
	theme  theme.Theme
	runner upbridge.Runner
	mode   mode
	cursor int
	cloud  bool
	err    error
	steps  []string
	spin   spinner.Model
}

// New constructs a Down wizard.
func New(ctx context.Context, t theme.Theme) Model {
	return Model{
		ctx:    ctx,
		theme:  t,
		runner: upbridge.DefaultRunner(),
		mode:   modeSelectCloud,
		spin:   pluralspinner.New(t),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case destroyDoneMsg:
		return m.applyDestroyDone(msg)
	case spinner.TickMsg:
		if m.mode != modeDestroying {
			return m, nil
		}
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	}
	return m, nil
}

func (m Model) updateKey(key tea.KeyPressMsg) (Model, tea.Cmd) {
	action := actionForKeystroke(key.Keystroke())
	switch m.mode {
	case modeSelectCloud:
		return m.updateSelectCloud(action, key)
	case modeAffirm:
		return m.updateAffirm(action, key)
	case modeDestroying:
		if action == keyActionBack {
			return m, nil // terraform running; ignore
		}
		return m, nil
	case modeComplete:
		if action == keyActionBack {
			return m, navigation.Navigate(navigation.Welcome)
		}
		return m, nil
	}
	return m, nil
}

func (m Model) updateSelectCloud(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	opts := cloudOptions()
	switch action {
	case keyActionBack:
		return m, navigation.Navigate(navigation.Welcome)
	case keyActionUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case keyActionDown:
		if m.cursor < len(opts)-1 {
			m.cursor++
		}
		return m, nil
	case keyActionConfirm:
		return m.selectCloud(opts[m.cursor])
	}
	text := keyText(key)
	for i, o := range opts {
		if text == cloudShortcut(o.id) || text == string(rune('1'+i)) {
			m.cursor = i
			return m.selectCloud(o)
		}
	}
	return m, nil
}

func (m Model) selectCloud(opt cloudOption) (Model, tea.Cmd) {
	m.cloud = opt.cloud
	m.err = nil
	if v, ok := utils.GetEnvBoolValue("PLURAL_DOWN_AFFIRM_DESTROY"); ok {
		if !v {
			m.err = fmt.Errorf("cancelled destroy")
			return m, nil
		}
		return m.beginDestroy()
	}
	m.mode = modeAffirm
	m.cursor = 0
	return m, nil
}

func (m Model) updateAffirm(action keyAction, key tea.KeyPressMsg) (Model, tea.Cmd) {
	opts := affirmOptions()
	switch action {
	case keyActionBack:
		m.mode = modeSelectCloud
		m.err = nil
		m.cursor = 0
		if m.cloud {
			m.cursor = 1
		}
		return m, nil
	case keyActionUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case keyActionDown:
		if m.cursor < len(opts)-1 {
			m.cursor++
		}
		return m, nil
	case keyActionConfirm:
		if !opts[m.cursor].value {
			m.err = fmt.Errorf("cancelled destroy")
			m.mode = modeSelectCloud
			m.cursor = 0
			if m.cloud {
				m.cursor = 1
			}
			return m, nil
		}
		return m.beginDestroy()
	}
	text := keyText(key)
	switch text {
	case "y", "Y":
		m.cursor = 0
		return m.beginDestroy()
	case "n", "N":
		m.err = fmt.Errorf("cancelled destroy")
		m.mode = modeSelectCloud
		m.cursor = 0
		if m.cloud {
			m.cursor = 1
		}
		return m, nil
	}
	return m, nil
}

func (m Model) beginDestroy() (Model, tea.Cmd) {
	m.mode = modeDestroying
	m.err = nil
	m.steps = nil
	runner := m.runner
	if runner == nil {
		runner = upbridge.DefaultRunner()
	}
	ctx := m.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	in := upbridge.DestroyInput{Cloud: m.cloud}
	if shouldExecLiveRunner(runner) {
		return m, destroyExecCmd(ctx, runner, in)
	}
	return m, tea.Batch(m.spin.Tick, m.destroyCmd(in))
}

func (m Model) destroyCmd(in upbridge.DestroyInput) tea.Cmd {
	runner := m.runner
	if runner == nil {
		runner = upbridge.DefaultRunner()
	}
	ctx := m.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return func() tea.Msg {
		var steps []string
		err := runner.Destroy(ctx, in, func(step string) {
			steps = append(steps, step)
		})
		return destroyDoneMsg{err: err, steps: steps}
	}
}

func (m Model) applyDestroyDone(msg destroyDoneMsg) (Model, tea.Cmd) {
	m.mode = modeComplete
	m.steps = msg.steps
	m.err = msg.err
	return m, nil
}

func cloudShortcut(id string) string {
	switch id {
	case "self-hosted":
		return "s"
	case "cloud":
		return "c"
	default:
		return ""
	}
}

func keyText(key tea.KeyPressMsg) string {
	text := key.Text
	if text == "" && key.Code > 0 && key.Code < 128 {
		text = string(rune(key.Code))
	}
	return text
}

// AffirmMessage is the same prompt plural down uses (for views/tests).
func AffirmMessage() string { return common.AffirmDown }

// Cloud reports whether --cloud was selected.
func (m Model) Cloud() bool { return m.cloud }

// ModeName is a test helper.
func (m Model) ModeName() string {
	switch m.mode {
	case modeSelectCloud:
		return "select-cloud"
	case modeAffirm:
		return "affirm"
	case modeDestroying:
		return "destroying"
	case modeComplete:
		return "complete"
	default:
		return "unknown"
	}
}
