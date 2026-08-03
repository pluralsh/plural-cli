package up

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"

	upbridge "github.com/pluralsh/plural-cli/pkg/bridge/up"
	"github.com/pluralsh/plural-cli/pkg/provider"
	"github.com/pluralsh/plural-cli/tui/navigation"
	"github.com/pluralsh/plural-cli/tui/theme"
)

func (f fakeProber) Probe(_ context.Context, providerID string) (upbridge.ProbeResult, error) {
	if f.err != nil {
		return upbridge.ProbeResult{}, f.err
	}
	fields := upbridge.ProviderFormFields(providerID)
	for i := range fields {
		switch fields[i].Key {
		case "region":
			fields[i].Options = []string{"us-east-2", "eu-west-1", "ap-southeast-1"}
		case "location":
			fields[i].Options = []string{"eastus", "westeurope"}
		case "project":
			fields[i].Options = []string{"demo-project", "other-project"}
		case "resourceGroup":
			fields[i].Options = []string{"rg-demo", provider.CreateNewOption}
		case "storageAccount":
			fields[i].Options = []string{"stordemo", provider.CreateNewOption}
		}
	}
	return upbridge.ProbeResult{
		Summary: "fake credentials ok · account 123456789012",
		Fields:  fields,
	}, nil
}

func (fakeProber) FieldOptions(_ context.Context, _, fieldKey string, values map[string]string) ([]string, error) {
	if fieldKey == "region" && values["project"] != "" {
		return []string{"us-east1", "europe-west1"}, nil
	}
	return nil, nil
}

func (f fakeProber) Preflights(context.Context, string, map[string]string) error {
	return f.preflightErr
}

type fakeProber struct {
	err          error
	preflightErr error
}

func testModel(t *testing.T) Model {
	t.Helper()
	model := NewWithProber(t.Context(), theme.New(colorprofile.ASCII), fakeProber{})
	model.gitChecker = func() bool { return true }
	model.domainLoader = func() domainMsg { return domainMsg{text: true} }
	return model
}

func testModelOutsideGit(t *testing.T, prober upbridge.Prober) Model {
	t.Helper()
	model := NewWithProber(t.Context(), theme.New(colorprofile.ASCII), prober)
	model.gitChecker = func() bool { return false }
	model.domainLoader = func() domainMsg { return domainMsg{text: true} }
	return model
}

func drainProbe(t *testing.T, model Model, _ tea.Cmd) Model {
	t.Helper()
	if model.mode != modeProbing {
		return model
	}
	msg := model.probeCmd()()
	model, _ = model.Update(msg)
	if model.mode == modeProbing {
		t.Fatalf("still probing after probeMsg: %#v", msg)
	}
	return model
}

func drainPreflight(t *testing.T, model Model) Model {
	t.Helper()
	if model.mode != modeRunPreflights {
		return model
	}
	msg := model.preflightCmd()()
	model, _ = model.Update(msg)
	return model
}

func drainDomain(t *testing.T, model Model) Model {
	t.Helper()
	if model.mode != modeAppDomain {
		return model
	}
	msg := model.loadDomainCmd()()
	model, _ = model.Update(msg)
	return model
}

func finishToSelected(t *testing.T, model Model) Model {
	t.Helper()
	model = drainPreflight(t, model)
	if model.mode == modeIgnoreContinue {
		model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	}
	if model.mode == modeSetupGit {
		model, _ = model.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	}
	if model.mode == modeSelectSCM {
		model, _ = model.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})
	}
	model = drainDomain(t, model)
	if model.mode == modeAppDomain {
		model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	}
	if model.mode == modeAffirmDeploy {
		model, _ = model.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	}
	if model.mode != modeSelected {
		t.Fatalf("expected selected, got %d err=%v", model.mode, model.err)
	}
	return model
}

func selectAWSForm(t *testing.T) Model {
	t.Helper()
	model := testModel(t)
	model, _ = model.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	model, _ = model.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
	model, cmd := model.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	model = drainProbe(t, model, cmd)
	if model.mode != modeProviderForm || model.SelectedProvider() != "aws" {
		t.Fatalf("after aws probe = mode=%d provider=%q err=%v", model.mode, model.SelectedProvider(), model.err)
	}
	return model
}

func TestSelfHostedProviderFormFlow(t *testing.T) {
	model := testModel(t)
	if model.mode != modeSelectFlow {
		t.Fatalf("mode = %d", model.mode)
	}

	model, _ = model.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	if model.mode != modeIgnorePreflights || model.SelectedFlow() != "self-hosted" {
		t.Fatalf("after self-hosted = mode=%d flow=%q", model.mode, model.SelectedFlow())
	}

	model, _ = model.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
	if model.mode != modeSelectProvider || !model.IgnorePreflights() {
		t.Fatalf("after ignore = mode=%d ignore=%v", model.mode, model.IgnorePreflights())
	}

	model, cmd := model.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	if model.mode != modeProbing {
		t.Fatalf("expected probing, got %d", model.mode)
	}
	model = drainProbe(t, model, cmd)
	if model.mode != modeProviderForm || model.SelectedProvider() != "aws" {
		t.Fatalf("after aws = mode=%d provider=%q", model.mode, model.SelectedProvider())
	}
	if model.credSummary == "" {
		t.Fatal("expected credential summary")
	}
	if model.formValues["region"] != "us-east-2" {
		t.Fatalf("default region = %q", model.formValues["region"])
	}
	// cluster is text; region is select
	model.formInput.SetValue("demo")
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // save cluster → region select
	if !model.currentIsSelect() {
		t.Fatalf("expected region select, field=%q", model.currentFieldKey())
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // pick us-east-2 → git/domain
	model = finishToSelected(t, model)
	view := model.View(80, 28)
	if !strings.Contains(view, "demo") || !strings.Contains(view, "--ignore-preflights") {
		t.Fatalf("plan view:\n%s", view)
	}
	if !strings.Contains(view, "fake credentials") {
		t.Fatalf("plan missing creds:\n%s", view)
	}

	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if model.mode != modeAffirmDeploy {
		t.Fatalf("esc to deploy affirm = %d", model.mode)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if model.mode != modeAppDomain {
		t.Fatalf("esc to domain = %d", model.mode)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if model.mode != modeProviderForm {
		t.Fatalf("esc to form = %d", model.mode)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if model.mode != modeSelectProvider {
		t.Fatalf("esc to providers = %d", model.mode)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if model.mode != modeIgnorePreflights {
		t.Fatalf("esc to preflights = %d", model.mode)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if model.mode != modeSelectFlow {
		t.Fatalf("esc to flows = %d", model.mode)
	}
	_, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil || cmd() != (navigation.NavigateMsg{Route: navigation.Welcome}) {
		t.Fatalf("expected welcome navigation")
	}
}

func TestProbeFailureBlocksWithoutIgnore(t *testing.T) {
	model := testModelOutsideGit(t, fakeProber{err: context.DeadlineExceeded})
	model, _ = model.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	model, _ = model.Update(tea.KeyPressMsg{Code: 'r', Text: "r"}) // run checks
	model, cmd := model.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	model = drainProbe(t, model, cmd)
	if model.mode != modeSelectProvider || model.err == nil {
		t.Fatalf("expected provider list with error, mode=%d err=%v", model.mode, model.err)
	}
}

func TestProbeFailureContinuesToGitAffirmWithIgnore(t *testing.T) {
	model := testModelOutsideGit(t, fakeProber{err: context.DeadlineExceeded})
	model, _ = model.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	model, _ = model.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
	model, cmd := model.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	model = drainProbe(t, model, cmd)
	if model.mode != modeIgnoreContinue {
		t.Fatalf("expected ignore-continue gate, got %d warn=%q", model.mode, model.probeWarn)
	}
	view := model.View(80, 24)
	if !strings.Contains(view, "continuing because --ignore-preflights") {
		t.Fatalf("missing ignore warning:\n%s", view)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model.mode != modeSetupGit {
		t.Fatalf("expected git affirm after enter, got %d", model.mode)
	}
	if len(model.formFields) != 0 {
		t.Fatalf("should not open region form, fields=%v", model.formFields)
	}
	if !strings.Contains(model.View(80, 28), "outside a git repository") {
		t.Fatalf("view:\n%s", model.View(80, 28))
	}

	model, _ = model.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	if model.mode != modeSelectSCM {
		t.Fatalf("after yes = %d err=%v", model.mode, model.err)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})
	if model.mode != modeAppDomain {
		t.Fatalf("after scm expected app domain, got %d", model.mode)
	}
	model = drainDomain(t, model)
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model.mode != modeAffirmDeploy {
		t.Fatalf("after domain expected deploy affirm, got %d", model.mode)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	if model.mode != modeSelected || model.scm.ID != "github" {
		t.Fatalf("after affirm = mode=%d scm=%q", model.mode, model.scm.ID)
	}
}

func TestPreflightFailureContinuesWithIgnore(t *testing.T) {
	model := testModel(t)
	model.prober = fakeProber{preflightErr: context.DeadlineExceeded}
	model, _ = model.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	model, _ = model.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
	model, cmd := model.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	model = drainProbe(t, model, cmd)
	model.formInput.SetValue("demo")
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model.mode != modeRunPreflights {
		t.Fatalf("expected run preflights, got %d", model.mode)
	}
	model = drainPreflight(t, model)
	if model.mode != modeIgnoreContinue || model.probeWarn == "" {
		t.Fatalf("expected ignore-continue after preflight fail, mode=%d warn=%q", model.mode, model.probeWarn)
	}
	if model.formValues["cluster"] != "demo" {
		t.Fatal("should keep form values when preflights fail with ignore")
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model.mode != modeSetupGit {
		t.Fatalf("expected git affirm after enter, got %d", model.mode)
	}
	model = finishToSelected(t, model)
}

func TestProbeFailureIgnoreInGitStillOpensAffirm(t *testing.T) {
	model := testModel(t)
	model.prober = fakeProber{err: context.DeadlineExceeded}
	model, _ = model.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	model, _ = model.Update(tea.KeyPressMsg{Code: 'i', Text: "i"})
	model, cmd := model.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	model = drainProbe(t, model, cmd)
	if model.mode != modeIgnoreContinue {
		t.Fatalf("expected ignore-continue, got %d", model.mode)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model.mode != modeSetupGit || !model.inGitRepo {
		t.Fatalf("expected in-repo git continue screen, mode=%d inGit=%v", model.mode, model.inGitRepo)
	}
	if !strings.Contains(model.View(80, 28), "Already inside a git work tree") {
		t.Fatalf("view:\n%s", model.View(80, 28))
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	if model.mode != modeAppDomain {
		t.Fatalf("yes should continue to app domain when already in git, got %d", model.mode)
	}
	model = drainDomain(t, model)
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if model.mode != modeAffirmDeploy {
		t.Fatalf("after domain expected deploy affirm, got %d", model.mode)
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	if model.mode != modeSelected {
		t.Fatalf("after affirm expected plan, got %d", model.mode)
	}
}

func TestPreflightFailureBlocksWithoutIgnore(t *testing.T) {
	model := testModel(t)
	model.prober = fakeProber{preflightErr: context.DeadlineExceeded}
	model, _ = model.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	model, _ = model.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	model, cmd := model.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	model = drainProbe(t, model, cmd)
	model.formInput.SetValue("demo")
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model = drainPreflight(t, model)
	if model.mode != modeProviderForm || model.err == nil {
		t.Fatalf("expected form with error, mode=%d err=%v", model.mode, model.err)
	}
}

func TestCloudModeAsksIgnorePreflights(t *testing.T) {
	model := testModel(t)
	model, _ = model.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	if model.mode != modeIgnorePreflights || !model.Cloud() {
		t.Fatalf("cloud = mode=%d cloud=%v", model.mode, model.Cloud())
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	if model.mode != modeCLITip || model.IgnorePreflights() {
		t.Fatalf("cli tip = mode=%d ignore=%v", model.mode, model.IgnorePreflights())
	}
	if !strings.Contains(model.View(80, 24), "plural up --cloud") || strings.Contains(model.View(80, 24), "--ignore-preflights") {
		t.Fatalf("cli tip:\n%s", model.View(80, 24))
	}
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if model.mode != modeIgnorePreflights {
		t.Fatalf("esc = %d", model.mode)
	}
}

func TestProviderFormValidation(t *testing.T) {
	model := selectAWSForm(t)
	model.formInput.SetValue("this-name-is-way-too-long")
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // to region
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // submit
	if model.mode != modeProviderForm || model.err == nil {
		t.Fatalf("expected validation error, mode=%d err=%v", model.mode, model.err)
	}
}

func TestFormThenGitAffirmOutsideRepo(t *testing.T) {
	model := testModelOutsideGit(t, fakeProber{})
	model, _ = model.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, cmd := model.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	model = drainProbe(t, model, cmd)
	model.formInput.SetValue("demo")
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	model = drainPreflight(t, model)
	if model.mode != modeSetupGit {
		t.Fatalf("expected git affirm after form, got %d", model.mode)
	}
}

func goldenModels(t *testing.T) (flow, ignore, provider, form, selected, cliTip, git, scm Model) {
	t.Helper()
	flow = testModel(t)
	ignore = flow
	ignore.mode = modeIgnorePreflights
	ignore.flow = ignore.flows[0]
	provider = ignore
	provider.mode = modeSelectProvider
	provider.ignoreAsked = true
	form, cmd := provider.beginProviderForm(provider.providers[0])
	form = drainProbe(t, form, cmd)
	form.formValues["cluster"] = "demo"
	form.applyLoadFormField()
	selected = form
	selected.mode = modeSelected
	selected.ignorePreflights = true
	selected.inGitRepo = true
	selected.appDomain = ""
	selected.formValues = map[string]string{"cluster": "demo", "region": "us-east-2"}
	cliTip = flow
	cliTip.mode = modeCLITip
	cliTip.flow = cliTip.flows[1]
	cliTip.ignorePreflights = true
	cliTip.ignoreAsked = true
	git = flow
	git.mode = modeSetupGit
	git.flow = git.flows[0]
	git.provider = git.providers[0]
	git.ignorePreflights = true
	git.probeWarn = "AWS credentials: failed"
	git.cursor = 0
	scm = git
	scm.mode = modeSelectSCM
	scm.probeWarn = ""
	return
}

func TestUpGoldens(t *testing.T) {
	flowModel, ignoreModel, providerModel, formModel, selected, cliTip, gitModel, scmModel := goldenModels(t)

	for _, tc := range []struct {
		name   string
		model  Model
		width  int
		height int
	}{
		{"flow-80", flowModel, 80, 24},
		{"flow-120", flowModel, 120, 30},
		{"ignore-80", ignoreModel, 80, 24},
		{"ignore-120", ignoreModel, 120, 30},
		{"provider-80", providerModel, 80, 24},
		{"provider-120", providerModel, 120, 30},
		{"form-80", formModel, 80, 24},
		{"form-120", formModel, 120, 30},
		{"selected-80", selected, 80, 24},
		{"selected-120", selected, 120, 30},
		{"clitip-80", cliTip, 80, 24},
		{"clitip-120", cliTip, 120, 30},
		{"git-80", gitModel, 80, 28},
		{"git-120", gitModel, 120, 30},
		{"scm-80", scmModel, 80, 24},
		{"scm-120", scmModel, 120, 30},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeView(tc.model.View(tc.width, tc.height))
			golden := filepath.Join("testdata", "up-"+tc.name+".golden")
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden: %v\nactual:\n%s", err, got)
			}
			if got != strings.TrimSuffix(string(want), "\n") {
				t.Fatalf("view changed\nwant:\n%s\n\ngot:\n%s", want, got)
			}
			lines := strings.Split(got, "\n")
			if len(lines) != tc.height {
				t.Fatalf("height = %d, want %d", len(lines), tc.height)
			}
			for _, line := range lines {
				if w := lipgloss.Width(line); w > tc.width {
					t.Fatalf("line width %d > %d: %q", w, tc.width, line)
				}
			}
		})
	}
}

func TestWriteUpGoldens(t *testing.T) {
	if os.Getenv("UPDATE_GOLDEN") == "" {
		t.Skip("set UPDATE_GOLDEN=1 to refresh fixtures")
	}
	flowModel, ignoreModel, providerModel, formModel, selected, cliTip, gitModel, scmModel := goldenModels(t)
	_ = os.MkdirAll("testdata", 0o755)
	for _, tc := range []struct {
		name   string
		model  Model
		width  int
		height int
	}{
		{"flow-80", flowModel, 80, 24},
		{"flow-120", flowModel, 120, 30},
		{"ignore-80", ignoreModel, 80, 24},
		{"ignore-120", ignoreModel, 120, 30},
		{"provider-80", providerModel, 80, 24},
		{"provider-120", providerModel, 120, 30},
		{"form-80", formModel, 80, 24},
		{"form-120", formModel, 120, 30},
		{"selected-80", selected, 80, 24},
		{"selected-120", selected, 120, 30},
		{"clitip-80", cliTip, 80, 24},
		{"clitip-120", cliTip, 120, 30},
		{"git-80", gitModel, 80, 28},
		{"git-120", gitModel, 120, 30},
		{"scm-80", scmModel, 80, 24},
		{"scm-120", scmModel, 120, 30},
	} {
		got := normalizeView(tc.model.View(tc.width, tc.height)) + "\n"
		if err := os.WriteFile(filepath.Join("testdata", "up-"+tc.name+".golden"), []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func normalizeView(view string) string {
	lines := strings.Split(ansi.Strip(view), "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	return strings.Join(lines, "\n")
}
