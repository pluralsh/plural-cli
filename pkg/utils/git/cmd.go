package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func gitRaw(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	res, err := execute(cmd)

	_ = fmt.Sprintf("cmds %s    res:\n %s \n", cmd.String(), res)

	return strings.TrimSpace(string(res)), err
}

func git(root string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	res, err := execute(cmd)
	return strings.TrimSpace(string(res)), err
}

func execute(cmd *exec.Cmd) (string, error) {
	res, err := cmd.CombinedOutput()
	if err != nil {
		return string(res), fmt.Errorf("Command %s failed with output:\n\n%s", cmd.String(), res)
	}

	return string(res), nil
}
