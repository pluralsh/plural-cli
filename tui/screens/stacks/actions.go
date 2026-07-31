package stacks

import (
	"strings"

	stacksbridge "github.com/pluralsh/plural-cli/pkg/bridge/stacks"
)

type actionKind uint8

const (
	actionGenBackend actionKind = iota
)

type detailAction struct {
	kind     actionKind
	shortcut string
	title    string
	blurb    string
}

func detailActions() []detailAction {
	return []detailAction{
		{kind: actionGenBackend, shortcut: "g", title: "Gen-backend", blurb: "write _override.tf · terraform backend"},
	}
}

type pendingOp struct {
	kind    actionKind
	title   string
	cli     string
	lines   []string
	backend *stacksbridge.GenBackendInput
}

func (m Model) genBackendPlan(input stacksbridge.GenBackendInput) pendingOp {
	d := m.detail
	dir := loCoalesce(strings.TrimSpace(input.Dir), ".")
	cli := "plural stacks gen-backend"
	if input.Address != "" {
		cli += " --address " + shellQuote(input.Address)
	}
	if input.LockAddress != "" {
		cli += " --lock-address " + shellQuote(input.LockAddress)
	}
	if input.UnlockAddress != "" {
		cli += " --unlock-address " + shellQuote(input.UnlockAddress)
	}
	lines := []string{
		"Action   Generate terraform backend override",
		"Stack    " + d.Name,
		"ID       " + d.ID,
		"Dir      " + dir,
		"File     _override.tf",
		"",
		"Writes _override.tf and appends it to .gitignore.",
		"Uses Console state URLs + your deploy token.",
	}
	if input.Address != "" || input.LockAddress != "" || input.UnlockAddress != "" {
		lines = append(lines,
			"",
			"Address  "+loCoalesce(input.Address, "(from Console)"),
			"Lock     "+loCoalesce(input.LockAddress, "(from Console)"),
			"Unlock   "+loCoalesce(input.UnlockAddress, "(from Console)"),
		)
	}
	return pendingOp{
		kind:    actionGenBackend,
		title:   "Gen-backend · " + d.Name,
		cli:     cli,
		backend: &input,
		lines:   lines,
	}
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

func formatResult(result stacksbridge.GenBackendResult) []string {
	return []string{
		"Wrote     " + result.FilePath,
		"Directory " + result.Dir,
		"",
		"Added _override.tf to .gitignore in that directory.",
	}
}
