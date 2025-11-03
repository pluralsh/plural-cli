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

func Preflight() (bool, error) {
	requirements := []string{"terraform", "git"}
	for _, req := range requirements {
		if ok, _ := utils.Which(req); !ok {
			return true, utils.HighlightError(fmt.Errorf("%s not installed", req))
		}
	}

	fmt.Print("\nTesting if git ssh is properly configured...\n")
	if err := checkGitSSH(); err != nil {
		fmt.Printf("%s\n\n", err.Error())
		utils.Warn("Please ensure that you have ssh keys set up for git and that you've added them to your ssh agent. You can use `plural crypto ssh-keygen` to create your first ssh keys then upload the public key to your git provider.\n")
		return true, fmt.Errorf("git ssh is not properly configured")
	}
	fmt.Println(" \033[32m (\u2713) \033[0m") // (✔)

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

	if strings.HasPrefix(url, "http") {
		utils.Error("Found non-ssh upstream url %s, please reclone the repo with SSH and retry.\n", url)
		utils.Warn("Please ensure that you have SSH keys set up for Git and that you've added them to your SSH agent.\n")
		utils.Warn("You can use `plural crypto ssh-keygen` to create your first SSH keys then upload the public key to your Git provider.\n")
		return true, fmt.Errorf("found non-ssh upstream")
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
		return fmt.Errorf("failed to clone: %s", b.String())
	}

	return nil
}
