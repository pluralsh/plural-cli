package common

import (
	"os"
	"os/exec"
)

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
