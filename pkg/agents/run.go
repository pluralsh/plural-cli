package agents

import (
	"strings"

	console "github.com/pluralsh/console/go/client"
)

func agentRunSessionURL(run *console.AgentRunMinimalFragment) string {
	if run == nil || run.GetUpload() == nil || run.GetUpload().GetSession() == nil {
		return ""
	}
	return strings.TrimSpace(*run.GetUpload().GetSession())
}

func agentRunPatchURL(run *console.AgentRunMinimalFragment) string {
	if run == nil || run.GetUpload() == nil || run.GetUpload().GetPatch() == nil {
		return ""
	}
	return strings.TrimSpace(*run.GetUpload().GetPatch())
}

func agentRunRef(run *console.AgentRunMinimalFragment) string {
	if run == nil {
		return ""
	}
	for _, pr := range run.GetPullRequests() {
		if pr == nil || pr.GetRef() == nil {
			continue
		}
		if ref := strings.TrimSpace(*pr.GetRef()); ref != "" {
			return ref
		}
	}
	return ""
}

func agentRunBranch(run *console.AgentRunMinimalFragment) string {
	if run == nil || run.GetBranch() == nil {
		return ""
	}
	return strings.TrimSpace(*run.GetBranch())
}

func agentRunProvider(run *console.AgentRunMinimalFragment) console.AgentRuntimeType {
	if run == nil || run.GetRuntime() == nil {
		return ""
	}
	return run.GetRuntime().Type
}
