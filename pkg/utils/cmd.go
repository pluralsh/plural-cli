package utils

import (
	"github.com/pluralsh/plural/pkg/config"
	"os"
	"fmt"
	"os/exec"
	"path/filepath"
)

func Cmd(conf *config.Config, program string, args ...string) error {
	return MkCmd(conf, program, args...).Run()
}

func Which(command string) (exists bool, path string) {
	root, _ := ProjectRoot()
	os.Setenv("PATH", fmt.Sprintf("%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))
	path, err := exec.LookPath(command)
	exists = err == nil
	return
}

func MkCmd(conf *config.Config, program string, args ...string) *exec.Cmd {
	cmd := exec.Command(program, args...)
	root, _ := ProjectRoot()
	os.Setenv("HELM_REPO_ACCESS_TOKEN", conf.Token)
	os.Setenv("PATH", fmt.Sprintf("%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func Install(command string, url string, dest string, postprocess func(string) (string, error)) error {
	if exists, _ := Which(command); exists {
		Success("%s is already installed\n", command)
		return nil
	}

	err := DownloadFile(url, dest)
	if err != nil {
		return err
	}

	bin, err := postprocess(dest)
	if err != nil {
		return err
	}

	if bin != "" {
		return os.Chmod(bin, 0777)
	}

	Success("%s successfully installed!\n", command)
	return nil
} 