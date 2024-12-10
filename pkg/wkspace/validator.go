package wkspace

import (
	"fmt"
	"os/exec"

	"github.com/pluralsh/plural-cli/pkg/utils"
)

func Preflight() (bool, error) {
	requirements := []string{"terraform", "git"}
	for _, req := range requirements {
		if ok, _ := utils.Which(req); !ok {
			return true, utils.HighlightError(fmt.Errorf("%s not installed", req))
		}
	}

	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	if _, err := cmd.CombinedOutput(); err != nil {
		return false, utils.HighlightError(fmt.Errorf("you're not in a git repository, you'll need to clone one before running plural"))
	}

	cmd = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	if _, err := cmd.CombinedOutput(); err != nil {
		return true, utils.HighlightError(fmt.Errorf("repository has no initial commit, you can simply commit a blank readme and push to start working"))
	}

	cmd = exec.Command("git", "ls-remote", "--exit-code")
	if _, err := cmd.CombinedOutput(); err != nil {
		return true, utils.HighlightError(fmt.Errorf("repository has no remotes set, make sure that at least one remote is set"))
	}

	return true, nil
}
