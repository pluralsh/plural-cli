package up

import (
	"os/exec"
)

// InGitRepo reports whether the current directory is inside a git work tree.
// Same check as wkspace.Preflight before the PLURAL_INIT_AFFIRM_SETUP_REPO prompt.
func InGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	return cmd.Run() == nil
}

// SetupGitPrompt is the Affirm message from HandleInitWithProject.
const SetupGitPrompt = "You're attempting to setup plural outside a git repository. Would you like us to set one up for you here?"
