// Package edge implements the TUI wizard for plural edge image, flash, bootstrap, and download.
package edge

import (
	"context"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	edgebridge "github.com/pluralsh/plural-cli/pkg/bridge/edge"
	pkgedge "github.com/pluralsh/plural-cli/pkg/edge"
	"github.com/pluralsh/plural-cli/tui/components/oplog"
	pluralspinner "github.com/pluralsh/plural-cli/tui/components/spinner"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type mode uint8

const (
	modeHub mode = iota
	modeImageForm
	modeImageReview
	modeImageRunning
	modeImageResult
	modeFlashForm
	modeFlashDevices
	modeFlashDevicePath
	modeFlashConfirm
	modeFlashRunning
	modeFlashResult
	modeBootstrapForm
	modeBootstrapReview
	modeBootstrapRunning
	modeBootstrapResult
	modeDownloadForm
	modeDownloadReview
	modeDownloadRunning
	modeDownloadResult
)

type field struct {
	key, label, placeholder string
	password                bool
	required                bool
}

type doneMsg struct {
	err     error
	request uint64
	kind    string
}

func imageFields() []field {
	return []field{
		{key: "output-dir", label: "Output directory", placeholder: "image"},
		{key: "project", label: "Project", placeholder: "default"},
		{key: "model", label: "Board model", placeholder: "rpi5"},
		{key: "username", label: "Username", placeholder: "plural"},
		{key: "password", label: "Password", placeholder: "required unless cloud-config is set", password: true},
		{key: "user", label: "User email", placeholder: "optional bootstrap token identity"},
		{key: "wifi-ssid", label: "Wi-Fi SSID", placeholder: "optional"},
		{key: "wifi-password", label: "Wi-Fi password", placeholder: "optional", password: true},
		{key: "cloud-config", label: "Cloud config path", placeholder: "optional; skips Console templating"},
		{key: "plural-config", label: "Plural config path", placeholder: "optional"},
		{key: "oci-url", label: "OCI push URL", placeholder: "optional"},
	}
}

func flashFields() []field {
	return []field{
		{key: "image", label: "Image file", placeholder: "path to kairos.img"},
		{key: "device", label: "Storage device", placeholder: "/dev/sdX"},
	}
}

func bootstrapFields() []field {
	return []field{
		{key: "machine-id", label: "Machine ID", placeholder: "unique id of the edge device", required: true},
		{key: "chart-loc", label: "Helm chart", placeholder: "optional URL or path"},
	}
}

func downloadFields() []field {
	return []field{
		{key: "oci-url", label: "OCI URL", placeholder: "required", required: true},
		{key: "to", label: "Destination", placeholder: "directory to unpack into"},
	}
}

func hubItems() []struct{ number, title, blurb string } {
	return []struct{ number, title, blurb string }{
		{"1", "image", "build Kairos ARM image (Docker)"},
		{"2", "flash", "write image onto a storage device"},
		{"3", "bootstrap", "register device and install agent"},
		{"4", "download", "pull and extract an OCI image"},
	}
}

// Model is the Edge screen: image build and flash.
type Model struct {
	ctx            context.Context
	loader         edgebridge.Loader
	theme          theme.Theme
	mode           mode
	cursor         int
	field          int
	inputs         []textinput.Model
	err            error
	needsAuth      bool
	request        uint64
	opLog          []string
	opLogCh        chan string
	opLogY         int
	opLogFollow    bool
	viewH          int
	viewW          int
	spin           spinner.Model
	exportDir      string
	logExportPath  string
	logExportErr   error
	listDevices    func() ([]pkgedge.FlashDevice, error)
	findImage      func() string
	devices        []pkgedge.FlashDevice
	deviceCursor   int
	deviceErr      error
	deviceCustom   bool
	imageSuggested bool
	flashWritten   int64
	flashTotal     int64
	flashStarted   time.Time
	progressCh     chan flashProgressMsg
}

func New(ctx context.Context, loader edgebridge.Loader, t theme.Theme) Model {
	return Model{
		ctx:         ctx,
		loader:      loader,
		theme:       t,
		spin:        pluralspinner.New(t),
		opLogFollow: true,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func newInputs(t theme.Theme, fields []field) []textinput.Model {
	styles := textinput.DefaultDarkStyles()
	styles.Focused.Text, styles.Focused.Prompt, styles.Focused.Placeholder = t.Body, t.Title, t.Muted
	styles.Blurred = styles.Focused
	inputs := make([]textinput.Model, len(fields))
	for i, field := range fields {
		input := textinput.New()
		input.Prompt = "› "
		input.Placeholder = field.placeholder
		input.CharLimit = 512
		input.SetStyles(styles)
		if field.password {
			input.EchoMode = textinput.EchoPassword
		}
		switch field.key {
		case "output-dir":
			input.SetValue("image")
		case "project":
			input.SetValue("default")
		case "model":
			input.SetValue("rpi5")
		case "username":
			input.SetValue("plural")
		case "to":
			if wd, err := os.Getwd(); err == nil {
				input.SetValue(wd)
			}
		}
		inputs[i] = input
	}
	if len(inputs) > 0 {
		inputs[0].Focus()
	}
	return inputs
}

func (m Model) startImage() Model {
	m.mode = modeImageForm
	m.err = nil
	m.needsAuth = false
	m.field = 0
	m.inputs = newInputs(m.theme, imageFields())
	return m
}

func (m Model) startFlash() Model {
	m.mode = modeFlashForm
	m.err = nil
	m.needsAuth = false
	m.field = 0
	m.inputs = newInputs(m.theme, flashFields())
	m.devices = nil
	m.deviceCursor = 0
	m.deviceErr = nil
	m.deviceCustom = false
	m.imageSuggested = false
	if path := m.defaultFlashImage(); path != "" && len(m.inputs) > 0 {
		m.inputs[0].SetValue(path)
		m.imageSuggested = true
	}
	return m
}

func (m Model) startBootstrap() Model {
	m.mode = modeBootstrapForm
	m.err = nil
	m.needsAuth = false
	m.field = 0
	m.inputs = newInputs(m.theme, bootstrapFields())
	return m
}

func (m Model) startDownload() Model {
	m.mode = modeDownloadForm
	m.err = nil
	m.needsAuth = false
	m.field = 0
	m.inputs = newInputs(m.theme, downloadFields())
	return m
}

func (m Model) openHubItem(index int) Model {
	switch index {
	case 1:
		return m.startFlash()
	case 2:
		return m.startBootstrap()
	case 3:
		return m.startDownload()
	default:
		return m.startImage()
	}
}

func (m Model) isRunning() bool {
	switch m.mode {
	case modeImageRunning, modeFlashRunning, modeBootstrapRunning, modeDownloadRunning:
		return true
	}
	return false
}

func (m Model) isForm() bool {
	switch m.mode {
	case modeImageForm, modeFlashForm, modeFlashDevicePath, modeBootstrapForm, modeDownloadForm:
		return true
	}
	return false
}

func (m Model) resultKind() string {
	switch m.mode {
	case modeFlashResult:
		return "flash"
	case modeBootstrapResult:
		return "bootstrap"
	case modeDownloadResult:
		return "download"
	default:
		return "image"
	}
}

func (m Model) updateReview(key tea.KeyPressMsg, back mode) (Model, tea.Cmd) {
	stroke := key.Keystroke()
	if stroke == "esc" {
		m.mode = back
		m.field = len(m.inputs) - 1
		if m.field >= 0 {
			m.inputs[m.field].Focus()
		}
		return m, nil
	}
	if stroke == "enter" {
		var cmd tea.Cmd
		switch back {
		case modeImageForm:
			cmd = m.beginImage()
		case modeBootstrapForm:
			cmd = m.beginBootstrap()
		case modeDownloadForm:
			cmd = m.beginDownload()
		}
		return m, cmd
	}
	return m, nil
}

func (m Model) defaultFlashImage() string {
	if m.findImage != nil {
		return m.findImage()
	}
	return pkgedge.DefaultFlashImage("")
}

func (m Model) imageOptions() pkgedge.ImageOptions {
	value := func(key string) string {
		for i, field := range imageFields() {
			if field.key == key && i < len(m.inputs) {
				return strings.TrimSpace(m.inputs[i].Value())
			}
		}
		return ""
	}
	return pkgedge.ImageOptions{
		OutputDir:    value("output-dir"),
		Project:      value("project"),
		User:         value("user"),
		PluralConfig: value("plural-config"),
		CloudConfig:  value("cloud-config"),
		Username:     value("username"),
		Password:     value("password"),
		WifiSSID:     value("wifi-ssid"),
		WifiPassword: value("wifi-password"),
		Model:        value("model"),
		OCIURL:       value("oci-url"),
	}
}

func (m Model) flashOptions() pkgedge.FlashOptions {
	value := func(key string) string {
		for i, field := range flashFields() {
			if field.key == key && i < len(m.inputs) {
				return strings.TrimSpace(m.inputs[i].Value())
			}
		}
		return ""
	}
	return pkgedge.FlashOptions{Image: value("image"), Device: value("device")}
}

func (m Model) bootstrapOptions() pkgedge.BootstrapOptions {
	value := func(key string) string {
		for i, field := range bootstrapFields() {
			if field.key == key && i < len(m.inputs) {
				return strings.TrimSpace(m.inputs[i].Value())
			}
		}
		return ""
	}
	return pkgedge.BootstrapOptions{MachineID: value("machine-id"), ChartLoc: value("chart-loc")}
}

func (m Model) downloadOptions() pkgedge.DownloadOptions {
	value := func(key string) string {
		for i, field := range downloadFields() {
			if field.key == key && i < len(m.inputs) {
				return strings.TrimSpace(m.inputs[i].Value())
			}
		}
		return ""
	}
	return pkgedge.DownloadOptions{OCIURL: value("oci-url"), To: value("to")}
}

func (m *Model) prepareRun(kind string) {
	switch kind {
	case "flash":
		m.mode = modeFlashRunning
	case "bootstrap":
		m.mode = modeBootstrapRunning
	case "download":
		m.mode = modeDownloadRunning
	default:
		m.mode = modeImageRunning
	}
	m.err = nil
	m.needsAuth = false
	m.opLog = nil
	m.opLogFollow = true
	m.opLogY = 0
	m.logExportPath = ""
	m.logExportErr = nil
	m.flashWritten = 0
	m.flashTotal = 0
	m.flashStarted = time.Time{}
	m.progressCh = nil
	m.request++
}

func (m *Model) beginImage() tea.Cmd {
	m.prepareRun("image")
	lines := make(chan string, 4096)
	m.opLogCh = lines
	return tea.Batch(m.spin.Tick, m.imageWorkCmd(lines), listenOpLog(lines))
}

func (m *Model) beginFlash() tea.Cmd {
	m.prepareRun("flash")
	lines := make(chan string, 4096)
	progress := make(chan flashProgressMsg, 8)
	m.opLogCh = lines
	m.progressCh = progress
	return tea.Batch(m.spin.Tick, m.flashWorkCmd(lines, progress), listenOpLog(lines), listenFlashProgress(progress))
}

func (m Model) imageWorkCmd(lines chan string) tea.Cmd {
	request, loader, ctx, options := m.request, m.loader, m.ctx, m.imageOptions()
	return func() tea.Msg {
		var err error
		if loader != nil {
			err = loader.BuildImage(ctx, options, func(line string) { sendLog(lines, line) })
		}
		if lines != nil {
			close(lines)
		}
		return doneMsg{err: err, request: request, kind: "image"}
	}
}

func (m Model) flashWorkCmd(lines chan string, progress chan flashProgressMsg) tea.Cmd {
	request, loader, ctx, options := m.request, m.loader, m.ctx, m.flashOptions()
	options.OnProgress = func(written, total int64) {
		sendFlashProgress(progress, flashProgressMsg{written: written, total: total})
	}
	return func() tea.Msg {
		var err error
		if loader != nil {
			err = loader.Flash(ctx, options, func(line string) { sendLog(lines, line) })
		}
		if lines != nil {
			close(lines)
		}
		if progress != nil {
			close(progress)
		}
		return doneMsg{err: err, request: request, kind: "flash"}
	}
}

func (m *Model) beginBootstrap() tea.Cmd {
	m.prepareRun("bootstrap")
	lines := make(chan string, 4096)
	m.opLogCh = lines
	return tea.Batch(m.spin.Tick, m.bootstrapWorkCmd(lines), listenOpLog(lines))
}

func (m *Model) beginDownload() tea.Cmd {
	m.prepareRun("download")
	lines := make(chan string, 4096)
	m.opLogCh = lines
	return tea.Batch(m.spin.Tick, m.downloadWorkCmd(lines), listenOpLog(lines))
}

func (m Model) bootstrapWorkCmd(lines chan string) tea.Cmd {
	request, loader, ctx, options := m.request, m.loader, m.ctx, m.bootstrapOptions()
	return func() tea.Msg {
		var err error
		if loader != nil {
			err = loader.Bootstrap(ctx, options, func(line string) { sendLog(lines, line) })
		}
		if lines != nil {
			close(lines)
		}
		return doneMsg{err: err, request: request, kind: "bootstrap"}
	}
}

func (m Model) downloadWorkCmd(lines chan string) tea.Cmd {
	request, loader, ctx, options := m.request, m.loader, m.ctx, m.downloadOptions()
	return func() tea.Msg {
		var err error
		if loader != nil {
			err = loader.Download(ctx, options, func(line string) { sendLog(lines, line) })
		}
		if lines != nil {
			close(lines)
		}
		return doneMsg{err: err, request: request, kind: "download"}
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case doneMsg:
		if msg.request != m.request {
			return m, nil
		}
		m.err = msg.err
		m.needsAuth = bridge.IsCode(msg.err, bridge.ErrorUnauthenticated)
		m.opLogCh = nil
		m.progressCh = nil
		m.opLogFollow = true
		switch msg.kind {
		case "flash":
			m.mode = modeFlashResult
		case "bootstrap":
			m.mode = modeBootstrapResult
		case "download":
			m.mode = modeDownloadResult
		default:
			m.mode = modeImageResult
		}
		if msg.err != nil {
			m.saveLogs(msg.kind)
		}
		return m, nil
	case opLogLineMsg:
		m.opLog = appendOpLog(m.opLog, msg.line)
		if m.opLogCh != nil {
			return m, tea.Batch(m.spin.Tick, listenOpLog(m.opLogCh))
		}
		return m, nil
	case flashProgressMsg:
		if m.flashStarted.IsZero() {
			m.flashStarted = time.Now()
		}
		m.flashWritten, m.flashTotal = msg.written, msg.total
		cmds := []tea.Cmd{m.spin.Tick}
		if m.progressCh != nil {
			cmds = append(cmds, listenFlashProgress(m.progressCh))
		}
		return m, tea.Batch(cmds...)
	case tea.WindowSizeMsg:
		m.viewH = msg.Height
		m.viewW = msg.Width
		return m, nil
	case spinner.TickMsg:
		if !m.isRunning() {
			return m, nil
		}
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		return m, cmd
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	}
	if m.isForm() && len(m.inputs) > 0 {
		var cmd tea.Cmd
		m.inputs[m.field], cmd = m.inputs[m.field].Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) updateKey(key tea.KeyPressMsg) (Model, tea.Cmd) {
	stroke := key.Keystroke()
	switch m.mode {
	case modeHub:
		if stroke == "esc" {
			return m, navigation.Navigate(navigation.Welcome)
		}
		if m.needsAuth && stroke == "c" {
			return m, navigation.Navigate(navigation.Access)
		}
		switch stroke {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(hubItems())-1 {
				m.cursor++
			}
		case "enter":
			return m.openHubItem(m.cursor), nil
		case "1":
			return m.startImage(), nil
		case "2":
			return m.startFlash(), nil
		case "3":
			return m.startBootstrap(), nil
		case "4":
			return m.startDownload(), nil
		}
		return m, nil
	case modeImageForm:
		return m.updateForm(key, imageFields(), modeHub, modeImageReview)
	case modeImageReview:
		return m.updateReview(key, modeImageForm)
	case modeBootstrapForm:
		return m.updateForm(key, bootstrapFields(), modeHub, modeBootstrapReview)
	case modeBootstrapReview:
		return m.updateReview(key, modeBootstrapForm)
	case modeDownloadForm:
		return m.updateForm(key, downloadFields(), modeHub, modeDownloadReview)
	case modeDownloadReview:
		return m.updateReview(key, modeDownloadForm)
	case modeFlashForm:
		return m.updateFlashImage(key)
	case modeFlashDevices:
		return m.updateFlashDevices(key)
	case modeFlashDevicePath:
		return m.updateFlashDevicePath(key)
	case modeFlashConfirm:
		if stroke == "esc" {
			if m.deviceCustom {
				m.mode = modeFlashDevicePath
				m.field = 1
				if len(m.inputs) > 1 {
					m.inputs[1].Focus()
				}
				return m, nil
			}
			m.mode = modeFlashDevices
			return m, nil
		}
		if stroke == "enter" {
			opts := m.flashOptions()
			if opts.Image == "" || opts.Device == "" {
				return m, nil
			}
			return m, m.beginFlash()
		}
		return m, nil
	case modeImageRunning, modeFlashRunning, modeBootstrapRunning, modeDownloadRunning:
		m.handleOpLogScroll(stroke)
		return m, nil
	case modeImageResult, modeFlashResult, modeBootstrapResult, modeDownloadResult:
		if m.handleOpLogScroll(stroke) {
			return m, nil
		}
		if stroke == "e" {
			m.saveLogs(m.resultKind())
			return m, nil
		}
		if m.needsAuth && stroke == "c" {
			return m, navigation.Navigate(navigation.Access)
		}
		if stroke == "esc" || stroke == "enter" {
			m.mode = modeHub
			m.err = nil
			m.needsAuth = false
			m.opLog = nil
			m.opLogFollow = true
			m.opLogY = 0
			m.logExportPath = ""
			m.logExportErr = nil
		}
		return m, nil
	}
	return m, nil
}

func (m Model) updateForm(key tea.KeyPressMsg, fields []field, back, next mode) (Model, tea.Cmd) {
	stroke := key.Keystroke()
	switch stroke {
	case "esc":
		if m.field == 0 {
			m.mode = back
			m.inputs = nil
			return m, nil
		}
		m.inputs[m.field].Blur()
		m.field--
		m.inputs[m.field].Focus()
		return m, nil
	case "enter":
		if fields[m.field].required && strings.TrimSpace(m.inputs[m.field].Value()) == "" {
			return m, nil
		}
		m.inputs[m.field].Blur()
		if m.field+1 < len(fields) {
			m.field++
			m.inputs[m.field].Focus()
			return m, nil
		}
		if next == modeFlashConfirm {
			opts := m.flashOptions()
			if opts.Image == "" || opts.Device == "" {
				m.inputs[m.field].Focus()
				return m, nil
			}
		}
		m.mode = next
		return m, nil
	}
	if len(m.inputs) == 0 {
		return m, nil
	}
	var cmd tea.Cmd
	m.inputs[m.field], cmd = m.inputs[m.field].Update(key)
	return m, cmd
}

func (m Model) updateFlashImage(key tea.KeyPressMsg) (Model, tea.Cmd) {
	stroke := key.Keystroke()
	switch stroke {
	case "esc":
		m.mode = modeHub
		m.inputs = nil
		return m, nil
	case "enter":
		if strings.TrimSpace(m.inputs[0].Value()) == "" {
			return m, nil
		}
		m.inputs[0].Blur()
		return m.showDevices(), nil
	}
	var cmd tea.Cmd
	m.inputs[0], cmd = m.inputs[0].Update(key)
	return m, cmd
}

func (m Model) showDevices() Model {
	m.mode = modeFlashDevices
	m.deviceCustom = false
	m.deviceCursor = 0
	list := m.listDevices
	if list == nil {
		list = pkgedge.ListFlashDevices
	}
	m.devices, m.deviceErr = list()
	return m
}

func (m Model) flashDeviceCount() int {
	return len(m.devices) + 1
}

func (m Model) updateFlashDevices(key tea.KeyPressMsg) (Model, tea.Cmd) {
	stroke := key.Keystroke()
	n := m.flashDeviceCount()
	switch stroke {
	case "esc":
		m.mode = modeFlashForm
		m.field = 0
		if len(m.inputs) > 0 {
			m.inputs[0].Focus()
		}
		return m, nil
	case "up", "k":
		if m.deviceCursor > 0 {
			m.deviceCursor--
		}
		return m, nil
	case "down", "j":
		if m.deviceCursor < n-1 {
			m.deviceCursor++
		}
		return m, nil
	case "r":
		return m.showDevices(), nil
	case "enter":
		return m.pickFlashDevice(m.deviceCursor)
	}
	if len(stroke) == 1 && stroke[0] >= '1' && stroke[0] <= '9' {
		idx := int(stroke[0] - '1')
		if idx < n {
			return m.pickFlashDevice(idx)
		}
	}
	return m, nil
}

func (m Model) pickFlashDevice(index int) (Model, tea.Cmd) {
	if index < 0 || index >= m.flashDeviceCount() {
		return m, nil
	}
	if index == len(m.devices) {
		m.mode = modeFlashDevicePath
		m.deviceCustom = true
		m.field = 1
		if len(m.inputs) > 1 {
			m.inputs[1].SetValue("")
			m.inputs[1].Focus()
		}
		return m, nil
	}
	m.deviceCustom = false
	if len(m.inputs) > 1 {
		m.inputs[1].SetValue(m.devices[index].Path)
		m.inputs[1].Blur()
	}
	m.mode = modeFlashConfirm
	return m, nil
}

func (m Model) updateFlashDevicePath(key tea.KeyPressMsg) (Model, tea.Cmd) {
	stroke := key.Keystroke()
	switch stroke {
	case "esc":
		return m.showDevices(), nil
	case "enter":
		if strings.TrimSpace(m.inputs[1].Value()) == "" {
			return m, nil
		}
		m.inputs[1].Blur()
		m.mode = modeFlashConfirm
		return m, nil
	}
	var cmd tea.Cmd
	m.inputs[1], cmd = m.inputs[1].Update(key)
	return m, cmd
}

func (m Model) selectedFlashDevice() (pkgedge.FlashDevice, bool) {
	path := m.flashOptions().Device
	for _, device := range m.devices {
		if device.Path == path {
			return device, true
		}
	}
	return pkgedge.FlashDevice{}, false
}

func (m Model) flashNeedsRoot() bool {
	if picked, ok := m.selectedFlashDevice(); ok {
		return picked.NeedsRoot
	}
	path := m.flashOptions().Device
	if path == "" {
		return false
	}
	return !pkgedge.DeviceWritable(path)
}

func (m *Model) handleOpLogScroll(stroke string) bool {
	window := 20
	if m.viewH > 0 {
		_, window = logPanelBudget(m.viewH, 4)
	}
	switch stroke {
	case "up", "k":
		m.scrollOpLog(-1, window)
		return true
	case "down", "j":
		m.scrollOpLog(1, window)
		return true
	case "pgup":
		m.scrollOpLog(-window, window)
		return true
	case "pgdown":
		m.scrollOpLog(window, window)
		return true
	case "home":
		m.opLogFollow = false
		m.opLogY = 0
		return true
	case "end":
		m.opLogFollow = true
		return true
	}
	return false
}

func (m *Model) scrollOpLog(delta, window int) {
	if window <= 0 {
		window = 20
	}
	total := len(wrapOpLog(m.opLog, opLogContentWidth(m.viewW)))
	maxStart := max(0, total-window)
	if m.opLogFollow {
		m.opLogY = maxStart
	}
	m.opLogFollow = false
	m.opLogY += delta
	if m.opLogY < 0 {
		m.opLogY = 0
	}
	if m.opLogY >= maxStart {
		m.opLogY = maxStart
		m.opLogFollow = true
	}
}

func (m *Model) saveLogs(kind string) {
	path, err := oplog.Write(m.exportDir, "edge-"+kind, m.opLog, m.err)
	m.logExportPath = path
	m.logExportErr = err
}
