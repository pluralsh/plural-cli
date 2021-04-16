package pluralfile

import (
	"fmt"
	"os"
	"os/exec"
)

type Command struct {
	Command string
	Args    []string
}

func (c *Command) Type() ComponentName {
	return COMMAND
}

func (c *Command) Key() string {
	return ""
}

func (c *Command) Push(repo string, sha string) (string, error) {
	fmt.Println("")
	cmd := exec.Command(c.Command, c.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return "", cmd.Run()
}
