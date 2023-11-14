package plural

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pluralsh/plural-cli/pkg/utils/git"
	"github.com/urfave/cli"
)

func handleRepair(c *cli.Context) error {
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	if err := git.Repair(repoRoot); err != nil {
		fmt.Println(err)
	}

	return nil
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
