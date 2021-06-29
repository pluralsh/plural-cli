package pluralfile

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("%s %s", c.Command, strings.Join(c.Args, " ")))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return "", cmd.Run()
}
