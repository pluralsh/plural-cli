package up

import (
	"os"
	"os/exec"

	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
)

func (ctx *Context) Prune() error {
	if ctx.Cloud {
		return nil
	}

	utils.Highlight("\nCleaning up unneeded resources...\n\n")
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	toRemove := []string{
		"null_resource.console",
		"helm_release.certmanager",
		"helm_release.flux",
		"helm_release.runtime",
		"helm_release.console",
	}

	for _, field := range toRemove {
		if err := stateRm("./terraform/mgmt", field); err != nil {
			return err
		}
	}

	if err := os.Remove("./terraform/mgmt/console.tf"); err != nil {
		return err
	}

	_ = os.RemoveAll("./terraform/apps")
	ctx.Cleanup()

	return git.Sync(repoRoot, "Post-setup resource cleanup", true)
}

func stateRm(dir, field string) error {
	cmd := exec.Command("terraform", "state", "rm", field)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
