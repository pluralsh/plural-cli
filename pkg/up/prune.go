package up

import (
	"os"
	"os/exec"

	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
)

func (ctx *Context) Prune() error {
	if ctx.Cloud {
		if err := ctx.runCheckpoint(ctx.Manifest.Checkpoint, "prune:cloud", func() error {
			return ctx.pruneCloud()
		}); err != nil {
			return err
		}
	}

	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	if err := ctx.runCheckpoint(ctx.Manifest.Checkpoint, "prune:mgmt", func() error {
		utils.Highlight("\nCleaning up unneeded resources...\n\n")

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
		return nil
	}); err != nil {
		return err
	}

	_ = os.Remove("./terraform/mgmt/console.tf")
	_ = os.RemoveAll("./terraform/apps")

	return git.Sync(repoRoot, "Post-setup resource cleanup", true)
}

func (ctx *Context) pruneCloud() error {
	utils.Highlight("\nCleaning up unneeded resources...\n\n")
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	_ = os.RemoveAll("./terraform/apps")

	return git.Sync(repoRoot, "Post-setup resource cleanup", true)
}

func stateRm(dir, field string) error {
	cmd := exec.Command("terraform", "state", "rm", field)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
