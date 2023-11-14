package server

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mitchellh/go-homedir"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

func serverFunc(f func(c *gin.Context) error) func(c *gin.Context) {
	return func(c *gin.Context) {
		if err := f(c); err != nil {
			_ = c.Error(err)
		}
	}
}

func toProvider(prov string) string {
	prov = strings.ToLower(prov)
	if prov == "gcp" {
		return "google"
	}
	return prov
}

func marker(name string) {
	if file, err := os.Create(markfile(name)); err == nil {
		file.Close()
	}
}

func marked(name string) bool {
	return utils.Exists(markfile(name))
}

func markfile(name string) string {
	p, _ := homedir.Expand("~/.plural")
	return filepath.Join(p, fmt.Sprintf("%s.mark", name))
}

func execCmd(command string, args ...string) error {
	_, err := execCmdWithOutput(command, args...)
	return err
}

func execCmdWithOutput(command string, args ...string) (string, error) {
	var buff bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.Stdout = &buff
	cmd.Stderr = &buff
	err := cmd.Run()
	return buff.String(), err
}
