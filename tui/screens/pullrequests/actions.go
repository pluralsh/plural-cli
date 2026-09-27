package pullrequests

import (
	"fmt"
	"sort"
	"strings"

	pullrequestsbridge "github.com/pluralsh/plural-cli/pkg/bridge/pullrequests"
)

type actionKind uint8

const (
	actionCreate actionKind = iota
	actionTrigger
	actionTemplate
	actionTest
	actionContracts
)

type detailAction struct {
	kind     actionKind
	shortcut string
	title    string
	blurb    string
	cliOnly  bool
}

func detailActions() []detailAction {
	return []detailAction{
		{kind: actionCreate, shortcut: "c", title: "Create", blurb: "open PR from automation"},
		{kind: actionTrigger, shortcut: "t", title: "Trigger", blurb: "name · configuration · branch"},
		{kind: actionTemplate, shortcut: "m", title: "Template", blurb: "local file · apply tree", cliOnly: true},
		{kind: actionTest, shortcut: "e", title: "Test", blurb: "local CRD", cliOnly: true},
		{kind: actionContracts, shortcut: "o", title: "Contracts", blurb: "contract suite", cliOnly: true},
	}
}

type pendingOp struct {
	kind    actionKind
	title   string
	cli     string
	lines   []string
	create  *pullrequestsbridge.CreatePRInput
	trigger *pullrequestsbridge.TriggerPRInput
}

func (m Model) createPlan(input pullrequestsbridge.CreatePRInput) pendingOp {
	d := m.detail
	cli := fmt.Sprintf("plural pr create %s", d.ID)
	if input.Branch != "" {
		cli += " --branch " + shellQuote(input.Branch)
	}
	if input.Context != "" {
		cli += " --context " + shellQuote(input.Context)
	}
	lines := []string{
		"Action      Create pull request",
		"Automation  " + d.Name,
		"ID          " + d.ID,
		"Branch      " + loCoalesce(input.Branch, "—"),
		"Context     " + loCoalesce(truncate(input.Context, 48), "—"),
	}
	return pendingOp{kind: actionCreate, title: "Create PR · " + d.Name, cli: cli, create: &input, lines: lines}
}

func (m Model) triggerPlan(input pullrequestsbridge.TriggerPRInput) pendingOp {
	d := m.detail
	cli := fmt.Sprintf("plural pr trigger %s", shellQuote(d.Name))
	if input.Branch != "" {
		cli += " --branch " + shellQuote(input.Branch)
	}
	keys := make([]string, 0, len(input.Configuration))
	for k := range input.Configuration {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	cfg := "—"
	if len(keys) > 0 {
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			v := input.Configuration[k]
			cli += " --configuration " + shellQuote(k+"="+v)
			parts = append(parts, k+"="+v)
		}
		cfg = strings.Join(parts, ", ")
	}
	lines := []string{
		"Action      Trigger PR automation",
		"Automation  " + d.Name,
		"Branch      " + loCoalesce(input.Branch, "—"),
		"Config      " + cfg,
	}
	return pendingOp{kind: actionTrigger, title: "Trigger · " + d.Name, cli: cli, trigger: &input, lines: lines}
}

func cliTipPlan(kind actionKind, name, file string) pendingOp {
	file = loCoalesce(strings.TrimSpace(file), "./automation.yaml")
	var title, cli string
	var lines []string
	switch kind {
	case actionTemplate:
		title = "Template · CLI"
		cli = fmt.Sprintf("plural pr template --file %s", shellQuote(file))
		lines = []string{
			"Action     Apply PR template in the local source tree",
			"File       " + file,
			"",
			"Local-only — TUI shows the CLI equivalent.",
		}
	case actionTest:
		title = "Test · CLI"
		cli = fmt.Sprintf("plural pr test --file %s", shellQuote(file))
		lines = []string{
			"Action     Test a PR automation CRD locally",
			"File       " + file,
			"",
			"Local-only — TUI shows the CLI equivalent.",
		}
	case actionContracts:
		title = "Contracts · CLI"
		cli = fmt.Sprintf("plural pr contracts --file %s", shellQuote(file))
		lines = []string{
			"Action     Run PR automation contract tests",
			"File       " + file,
			"",
			"Local-only — TUI shows the CLI equivalent.",
		}
	}
	_ = name
	return pendingOp{kind: kind, title: title, cli: cli, lines: lines}
}

func parseConfiguration(raw string) (map[string]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]string{}, nil
	}
	out := map[string]string{}
	for _, part := range strings.Fields(raw) {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 || strings.TrimSpace(kv[0]) == "" {
			return nil, fmt.Errorf("invalid configuration %q (expected key=value)", part)
		}
		out[kv[0]] = kv[1]
	}
	return out, nil
}

func shellQuote(v string) string {
	if v == "" {
		return `""`
	}
	if strings.ContainsAny(v, " \t\"'") {
		return `"` + strings.ReplaceAll(v, `"`, `\"`) + `"`
	}
	return v
}

func truncate(v string, n int) string {
	if len(v) <= n {
		return v
	}
	return v[:n-1] + "…"
}
