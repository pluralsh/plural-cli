package wkspace

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/executor"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/git"
)

func Preflight(dryRun, ignorePreflights bool) (bool, error) {
	requirements := []string{"terraform", "git"}
	if dryRun {
		requirements = []string{"git"}
	}

	fmt.Printf("Checking required CLI dependencies: %s\n", strings.Join(requirements, ", "))
	for _, req := range requirements {
		if ok, _ := utils.Which(req); !ok {
			return true, utils.HighlightError(fmt.Errorf("required CLI dependency %q is not installed or not found in $PATH", req))
		}
	}
	fmt.Print("All required CLI dependencies are present.\n\n")

	fmt.Print("Checking git repository setup...\n")
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

	url, err := git.GetURL()
	if err != nil {
		return true, err
	}

	isHTTPS := strings.HasPrefix(url, "http")

	if !dryRun && !ignorePreflights && !isHTTPS {
		fmt.Print("\nTesting if git ssh is properly configured...")
		if err := checkGitSSH(); err != nil {
			fmt.Printf("%s\n\n", err.Error())
			utils.Warn("Please ensure that you have ssh keys set up for git and that you've added them to your ssh agent. You can use `plural crypto ssh-keygen` to create your first ssh keys then upload the public key to your git provider.\n")
			return true, fmt.Errorf("git ssh is not properly configured")
		}
		fmt.Printf(" \033[32m (\u2713) \033[0m\n\n") // (✓)
	}

	return true, nil
}

func checkGitSSH() error {
	dir, err := os.MkdirTemp("", "scaffolds")
	if err != nil {
		return err
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			err := utils.HighlightError(err)
			if err != nil {
				return
			}
		}
	}(dir)

	cmd := exec.Command("git", "clone", "git@github.com:pluralsh/scaffolds.git", dir)
	var b bytes.Buffer
	// Configure the output to display progress as dots
	cmd.Stdout = &executor.OutputWriter{Delegate: os.Stdout}
	cmd.Stderr = &b
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("SSH connectivity test failed (cloning git@github.com:pluralsh/scaffolds.git): %s", b.String())
	}

	return nil
}
