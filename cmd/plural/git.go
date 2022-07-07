package main

import (
	"os"
	"os/exec"

	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/urfave/cli"
)

func handleRepair(c *cli.Context) error {
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	return git.Repair(repoRoot)
}

func gitConfig(name, val string) error {
	cmd := gitCommand("config", name, val)
	return cmd.Run()
}

func gitCommand(args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
