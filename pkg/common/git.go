package common

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pluralsh/plural-cli/pkg/utils/git"
	"github.com/urfave/cli"
)

func HandleRepair(c *cli.Context) error {
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	if err := git.Repair(repoRoot); err != nil {
		fmt.Println(err)
	}

	return nil
}

func GitConfig(name, val string) error {
	cmd := GitCommand("config", name, val)
	return cmd.Run()
}

func GitCommand(args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
