package up

import (
	"os"
	"os/exec"

	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
)

func (c *Context) Prune() error {
	if c.Cloud {
		return c.runCheckpoint(c.Manifest.Checkpoint, "prune:cloud", func() error {
			return c.pruneCloud()
		})
	}

	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	if err := c.runCheckpoint(c.Manifest.Checkpoint, "prune:mgmt", func() error {
		utils.Highlight("\nCleaning up unneeded resources...\n\n")

		toRemove := []string{
			"null_resource.console",
			"helm_release.certmanager",
			"helm_release.flux",
			"helm_release.runtime",
			"helm_release.console",
			"kubernetes_namespace.infra",
			"kubernetes_secret.runtime_config",
			"kubernetes_secret.console_config",
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
	_ = os.Remove("./terraform/mgmt/config_secrets.tf")
	_ = os.RemoveAll("./temp")
	_ = os.RemoveAll("./terraform/apps")
	_ = os.Remove("./context.yaml")

	return git.Sync(repoRoot, "Post-setup resource cleanup", true)
}

func (c *Context) pruneCloud() error {
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

// stateRmBestEffort is like stateRm but silently ignores the case where the
// resource was never in state (e.g. install_prereqs=false skipped certmanager/flux).
func stateRmBestEffort(dir, field string) {
	_ = stateRm(dir, field)
}

// pruneBYOK removes the bootstrap helm/null resources from terraform state and
// cleans up the one-shot files used during installation (no cloud infra to touch).
func (c *Context) pruneBYOK() error {
	return c.runCheckpoint(c.Manifest.Checkpoint, "prune:mgmt", func() error {
		utils.Highlight("\nCleaning up unneeded resources...\n\n")

		// These may or may not be in state depending on install_prereqs value.
		stateRmBestEffort("./terraform/mgmt", "helm_release.certmanager")
		stateRmBestEffort("./terraform/mgmt", "helm_release.flux")

		// These are always created for BYOK.
		required := []string{
			"null_resource.console",
			"helm_release.runtime",
			"helm_release.console",
		}
		for _, field := range required {
			if err := stateRm("./terraform/mgmt", field); err != nil {
				return err
			}
		}

		_ = os.Remove("./terraform/mgmt/console.tf")
		_ = os.Remove("./terraform/mgmt/config_secrets.tf")
		_ = os.RemoveAll("./temp")
		_ = os.RemoveAll("./terraform/apps")
		_ = os.Remove("./context.yaml")

		repoRoot, err := git.Root()
		if err != nil {
			return err
		}
		return git.Sync(repoRoot, "Post-setup resource cleanup", true)
	})
}
