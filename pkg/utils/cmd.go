package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pluralsh/plural-cli/pkg/utils/pathing"
)

func Exec(program string, args ...string) error {
	cmd := exec.Command(program, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Execute(cmd *exec.Cmd) error {
	res, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command %s failed with output:\n\n%s\n%w", cmd.String(), res, err)
	}

	return nil
}

func ExecuteWithOutput(cmd *exec.Cmd) (string, error) {
	res, err := cmd.CombinedOutput()
	if err != nil {
		return string(res), fmt.Errorf("command %s failed with output:\n\n%s", cmd.String(), res)
	}

	return string(res), nil
}

func Which(command string) (exists bool, path string) {
	root, _ := ProjectRoot()
	os.Setenv("PATH", fmt.Sprintf("%s:%s", pathing.SanitizeFilepath(filepath.Join(root, "bin")), os.Getenv("PATH")))
	path, err := exec.LookPath(command)
	exists = err == nil
	return
}
