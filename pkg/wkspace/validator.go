package wkspace

import (
	"fmt"
	"os/exec"

	"github.com/pluralsh/plural-cli/pkg/api"
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

func (wk *Workspace) providersValid(providers []string) error {
	if len(providers) == 0 {
		return nil
	}

	pass := false
	for _, provider := range providers {
		if wk.match(provider) {
			pass = true
		}
	}

	if !pass {
		return fmt.Errorf("provider %s is not supported for any of %v", wk.Provider.Name(), providers)
	}

	return nil
}

func (wk *Workspace) match(prov string) bool {
	switch wk.Provider.Name() {
	case api.ProviderGCP:
		return prov == "GCP"
	case api.ProviderAWS:
		return prov == "AWS"
	case api.ProviderAzure:
		return prov == "AZURE"
	default:
		return false
	}
}
