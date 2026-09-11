package edge

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	"github.com/pluralsh/plural-cli/pkg/bridge"
	pkgedge "github.com/pluralsh/plural-cli/pkg/edge"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

type fakeLoader struct {
	image pkgedge.ImageOptions
	flash pkgedge.FlashOptions
}

func (f *fakeLoader) BuildImage(_ context.Context, options pkgedge.ImageOptions, log func(string)) error {
	f.image = options
	if log != nil {
		log("reading configuration")
		log("preparing output directory")
	}
	return nil
}
func (f *fakeLoader) Flash(_ context.Context, options pkgedge.FlashOptions, log func(string)) error {
	f.flash = options
	if log != nil {
		log("flashing " + options.Image + " onto " + options.Device)
	}
	return nil
}

func drainWork(t *testing.T, model Model) Model {
	t.Helper()
	var msg tea.Msg
	switch model.mode {
	case modeImageRunning:
		msg = model.imageWorkCmd(nil)()
	case modeFlashRunning:
		msg = model.flashWorkCmd(nil, nil)()
	default:
		t.Fatalf("expected running, got %d", model.mode)
	}
	model, _ = model.Update(msg)
	return model
}

func TestHubOpensImageAndFlash(t *testing.T) {
	model := New(t.Context(), &fakeLoader{}, theme.New(colorprofile.ASCII))
	got := normalizeView(model.View(80, 24))
	if !strings.Contains(got, "Edge commands") || !strings.Contains(got, "image") || !strings.Contains(got, "flash") {
		t.Fatalf("hub missing commands:\n%s", got)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
	if model.mode != modeImageForm {
		t.Fatalf("mode = %d", model.mode)
	}
	got = normalizeView(model.View(80, 24))
	if !strings.Contains(got, "plural edge image") {
		t.Fatalf("image form missing CLI hint:\n%s", got)
	}
	model = New(t.Context(), &fakeLoader{}, theme.New(colorprofile.ASCII))
	model, _ = model.Update(tea.KeyPressMsg{Code: '2', Text: "2"})
	if model.mode != modeFlashForm {
		t.Fatalf("mode = %d", model.mode)
	}
}

func TestFlashPrefillsLocalKairosImage(t *testing.T) {
	path := "/work/edge/image/kairos.img"
	model := New(t.Context(), &fakeLoader{}, theme.New(colorprofile.ASCII))
	model.findImage = func() string { return path }
	model, _ = model.Update(tea.KeyPressMsg{Code: '2', Text: "2"})
	if model.inputs[0].Value() != path || !model.imageSuggested {
		t.Fatalf("image = %q suggested=%v", model.inputs[0].Value(), model.imageSuggested)
	}
	got := normalizeView(model.View(80, 24))
	if !strings.Contains(got, path) || !strings.Contains(got, "Found image/kairos.img") {
		t.Fatalf("form missing suggested image:\n%s", got)
	}
}

func TestImageReviewQueuesBuild(t *testing.T) {
	loader := &fakeLoader{}
	model := New(t.Context(), loader, theme.New(colorprofile.ASCII))
	model, _ = model.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
	for i := 0; i < len(imageFields()); i++ {
		if imageFields()[i].key == "password" {
			model.inputs[i].SetValue("secret")
		}
		if imageFields()[i].key == "cloud-config" {
			model.inputs[i].SetValue("/tmp/cloud.yaml")
		}
		model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	}
	if model.mode != modeImageReview {
		t.Fatalf("expected review, got %d", model.mode)
	}
	got := normalizeView(model.View(80, 24))
	if !strings.Contains(got, "image") || strings.Contains(got, "secret") {
		t.Fatalf("review should hide password:\n%s", got)
	}
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("review did not start build")
	}
	if model.mode != modeImageRunning {
		t.Fatalf("expected running, got %d", model.mode)
	}
	got = normalizeView(model.View(80, 24))
	if !strings.Contains(got, "Running plural edge image") || !strings.Contains(got, "streams below") {
		t.Fatalf("running view should look like terraform logs:\n%s", got)
	}
	model = drainWork(t, model)
	if model.mode != modeImageResult || loader.image.Password != "secret" || loader.image.CloudConfig != "/tmp/cloud.yaml" {
		t.Fatalf("build options = %#v mode=%d", loader.image, model.mode)
	}
}

func TestFlashRequiresConfirm(t *testing.T) {
	loader := &fakeLoader{}
	model := New(t.Context(), loader, theme.New(colorprofile.ASCII))
	model.listDevices = func() ([]pkgedge.FlashDevice, error) {
		return []pkgedge.FlashDevice{{Path: "/dev/sdb", Model: "SanDisk Ultra", Size: 16 << 30}}, nil
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: '2', Text: "2"})
	model.inputs[0].SetValue("kairos.img")
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model.mode != modeFlashDevices {
		t.Fatalf("mode = %d", model.mode)
	}
	got := normalizeView(model.View(80, 24))
	if !strings.Contains(got, "/dev/sdb") || !strings.Contains(got, "SanDisk Ultra") {
		t.Fatalf("device picker missing USB disk:\n%s", got)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
	if model.mode != modeFlashConfirm {
		t.Fatalf("mode = %d", model.mode)
	}
	got = normalizeView(model.View(80, 24))
	if !strings.Contains(got, "overwrites") || !strings.Contains(got, "/dev/sdb") {
		t.Fatalf("confirm missing warning:\n%s", got)
	}
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("confirm did not start flash")
	}
	if model.mode != modeFlashRunning {
		t.Fatalf("expected running, got %d", model.mode)
	}
	model = drainWork(t, model)
	if loader.flash.Image != "kairos.img" || loader.flash.Device != "/dev/sdb" {
		t.Fatalf("flash options = %#v", loader.flash)
	}
}

func TestFlashConfirmWarnsNeedsRoot(t *testing.T) {
	model := New(t.Context(), &fakeLoader{}, theme.New(colorprofile.ASCII))
	model.listDevices = func() ([]pkgedge.FlashDevice, error) {
		return []pkgedge.FlashDevice{{Path: "/dev/sdb", Model: "SanDisk Ultra", Size: 16 << 30, NeedsRoot: true}}, nil
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: '2', Text: "2"})
	model.inputs[0].SetValue("kairos.img")
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	got := normalizeView(model.View(80, 24))
	if !strings.Contains(got, "needs root") {
		t.Fatalf("picker should mark needs root:\n%s", got)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
	got = normalizeView(model.View(80, 24))
	if !strings.Contains(got, "Needs root") || !strings.Contains(got, "sudo plural tui") {
		t.Fatalf("confirm should warn about root:\n%s", got)
	}
}

func TestFlashEmptyDeviceStaysOnForm(t *testing.T) {
	model := New(t.Context(), &fakeLoader{}, theme.New(colorprofile.ASCII))
	model.listDevices = func() ([]pkgedge.FlashDevice, error) { return nil, nil }
	model, _ = model.Update(tea.KeyPressMsg{Code: '2', Text: "2"})
	model.inputs[0].SetValue("kairos.img")
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model.mode != modeFlashDevices {
		t.Fatalf("mode = %d", model.mode)
	}
	got := normalizeView(model.View(80, 24))
	if !strings.Contains(got, "No USB disks found") || !strings.Contains(got, "Enter path") {
		t.Fatalf("empty USB list:\n%s", got)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model.mode != modeFlashDevicePath {
		t.Fatalf("mode = %d", model.mode)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model.mode != modeFlashDevicePath {
		t.Fatalf("empty device continued to %d", model.mode)
	}
}

func TestFlashCustomPath(t *testing.T) {
	loader := &fakeLoader{}
	model := New(t.Context(), loader, theme.New(colorprofile.ASCII))
	model.listDevices = func() ([]pkgedge.FlashDevice, error) {
		return []pkgedge.FlashDevice{{Path: "/dev/sdb"}}, nil
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: '2', Text: "2"})
	model.inputs[0].SetValue("kairos.img")
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(tea.KeyPressMsg{Code: '2', Text: "2"})
	if model.mode != modeFlashDevicePath {
		t.Fatalf("mode = %d", model.mode)
	}
	model.inputs[1].SetValue("/dev/sdc")
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model.mode != modeFlashConfirm || model.flashOptions().Device != "/dev/sdc" {
		t.Fatalf("custom path = %#v mode=%d", model.flashOptions(), model.mode)
	}
}

func TestImageUnauthenticatedOffersAccess(t *testing.T) {
	loader := &errLoader{err: &bridge.Error{Code: bridge.ErrorUnauthenticated, Err: errors.New("connect a Console profile")}}
	model := New(t.Context(), loader, theme.New(colorprofile.ASCII))
	model, _ = model.Update(tea.KeyPressMsg{Code: '1', Text: "1"})
	for i := 0; i < len(imageFields()); i++ {
		if imageFields()[i].key == "password" {
			model.inputs[i].SetValue("secret")
		}
		model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	}
	model, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("review did not start build")
	}
	model = drainWork(t, model)
	if !model.needsAuth || model.mode != modeImageResult {
		t.Fatalf("needsAuth=%v mode=%d", model.needsAuth, model.mode)
	}
	got := normalizeView(model.View(80, 24))
	if !strings.Contains(got, "Press c to open Access") {
		t.Fatalf("missing connect hint:\n%s", got)
	}
	_, cmd = model.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	if cmd == nil {
		t.Fatal("c did not navigate")
	}
	if got := cmd().(navigation.NavigateMsg).Route; got != navigation.Access {
		t.Fatalf("route = %s", got)
	}
}

type errLoader struct{ err error }

func (e *errLoader) BuildImage(context.Context, pkgedge.ImageOptions, func(string)) error {
	return e.err
}
func (e *errLoader) Flash(context.Context, pkgedge.FlashOptions, func(string)) error {
	return e.err
}

func TestHubEscReturnsToWelcome(t *testing.T) {
	model := New(t.Context(), &fakeLoader{}, theme.New(colorprofile.ASCII))
	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("esc did not navigate")
	}
}

func TestImageRunningKeepsLogsInPanel(t *testing.T) {
	model := New(t.Context(), &fakeLoader{}, theme.New(colorprofile.ASCII))
	model.mode = modeImageRunning
	model.viewH = 24
	model, _ = model.Update(opLogLineMsg{line: "writing plural-bundle bundle"})
	got := normalizeView(model.View(80, 24))
	if !strings.Contains(got, "Building") || !strings.Contains(got, "writing plural-bundle bundle") {
		t.Fatalf("expected terraform-style log panel:\n%s", got)
	}
}

func TestFlashRunningShowsProgressBar(t *testing.T) {
	model := New(t.Context(), &fakeLoader{}, theme.New(colorprofile.ASCII))
	model.mode = modeFlashRunning
	model.flashWritten = 6 << 30
	model.flashTotal = 10 << 30
	model.flashStarted = time.Now().Add(-3 * time.Second)
	model, _ = model.Update(flashProgressMsg{written: 6 << 30, total: 10 << 30})
	got := normalizeView(model.View(80, 24))
	if !strings.Contains(got, "Flashing") || !strings.Contains(got, "60%") {
		t.Fatalf("missing percent:\n%s", got)
	}
	if !strings.Contains(got, "█") || !strings.Contains(got, "░") {
		t.Fatalf("missing progress bar:\n%s", got)
	}
	if strings.Contains(got, "bytes (") || strings.Contains(got, "copied,") {
		t.Fatalf("raw dd output leaked into the flash window:\n%s", got)
	}
	if !strings.Contains(got, "6.0 GB") || !strings.Contains(got, "10.0 GB") {
		t.Fatalf("missing size label:\n%s", got)
	}
}

func TestImageResultKeepsLogs(t *testing.T) {
	model := New(t.Context(), &fakeLoader{}, theme.New(colorprofile.ASCII))
	model.exportDir = t.TempDir()
	model.mode = modeImageResult
	model.opLog = []string{"reading configuration", "image saved to image directory"}
	got := normalizeView(model.View(80, 24))
	if !strings.Contains(got, "Image complete") || !strings.Contains(got, "reading configuration") {
		t.Fatalf("result should keep logs:\n%s", got)
	}
	if !strings.Contains(got, "Scroll logs") {
		t.Fatalf("missing scroll hint:\n%s", got)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: 'e', Text: "e"})
	if model.logExportPath == "" || model.logExportErr != nil {
		t.Fatalf("export path=%q err=%v", model.logExportPath, model.logExportErr)
	}
}

func TestOpLogLinesUsePanelWidth(t *testing.T) {
	model := New(t.Context(), &fakeLoader{}, theme.New(colorprofile.ASCII))
	model.mode = modeImageResult
	long := strings.Repeat("abcdefghij", 20)
	model.opLog = []string{long}
	view := ansi.Strip(model.View(160, 24))
	if !strings.Contains(view, strings.Repeat("abcdefghij", 12)) {
		t.Fatalf("expected wrapped log to keep the start of the line, got:\n%s", view)
	}
	if strings.Contains(view, long) {
		t.Fatal("expected wrap so the 200-char line is not a single row")
	}
}

func normalizeView(view string) string {
	lines := strings.Split(ansi.Strip(view), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	return strings.Join(lines, "\n")
}
