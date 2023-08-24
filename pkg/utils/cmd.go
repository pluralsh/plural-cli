package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

func Cmd(conf *config.Config, program string, args ...string) error {
	return MkCmd(conf, program, args...).Run()
}

func Exec(program string, args ...string) error {
	cmd := exec.Command(program, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Execute(cmd *exec.Cmd) error {
	res, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Command %s failed with output:\n\n%s\n%s", cmd.String(), res, err)
	}

	return nil
}

func ExecuteWithOutput(cmd *exec.Cmd) (string, error) {
	res, err := cmd.CombinedOutput()
	if err != nil {
		return string(res), fmt.Errorf("Command %s failed with output:\n\n%s", cmd.String(), res)
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

func MkCmd(conf *config.Config, program string, args ...string) *exec.Cmd {
	cmd := exec.Command(program, args...)
	root, _ := ProjectRoot()
	os.Setenv("HELM_REPO_ACCESS_TOKEN", conf.Token)
	os.Setenv("PATH", fmt.Sprintf("%s:%s", pathing.SanitizeFilepath(filepath.Join(root, "bin")), os.Getenv("PATH")))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
