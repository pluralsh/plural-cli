package up

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/pluralsh/plural-cli/pkg/utils"
)

type terraformCmd struct {
	dir     string
	cmd     string
	args    []string
	retries int
}

func (ctx *Context) Deploy(commit func() error) error {
	if err := ctx.Provider.CreateBucket(); err != nil {
		return err
	}

	if err := runAll([]terraformCmd{
		{dir: "./clusters", cmd: "init", args: []string{"-upgrade"}},
		{dir: "./clusters", cmd: "apply", args: []string{"-auto-approve"}, retries: 1},
	}); err != nil {
		return err
	}

	if err := ping(fmt.Sprintf("https://console.%s", ctx.Manifest.Network.Subdomain)); err != nil {
		return err
	}

	if err := commit(); err != nil {
		return err
	}

	utils.Highlight("\nSetting up gitops management...\n")

	return runAll([]terraformCmd{
		{dir: "./apps/terraform", cmd: "init", args: []string{"-upgrade"}},
		{dir: "./apps/terraform", cmd: "apply", args: []string{"-auto-approve"}, retries: 1},
	})
}

func (ctx *Context) Destroy() error {
	return runAll([]terraformCmd{
		{dir: "./clusters", cmd: "init", args: []string{"-upgrade"}},
		{dir: "./clusters", cmd: "destroy", args: []string{"-auto-approve"}, retries: 2},
	})
}

func runAll(cmds []terraformCmd) error {
	for _, cmd := range cmds {
		if err := cmd.run(); err != nil {
			return err
		}
	}

	return nil
}

func (tf *terraformCmd) run() (err error) {
	for tf.retries >= 0 {
		args := append([]string{tf.cmd}, tf.args...)
		cmd := exec.Command("terraform", args...)
		cmd.Dir = tf.dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err == nil {
			return
		}

		tf.retries -= 1
		if tf.retries >= 0 {
			utils.Warn("terraform cmd failed, retrying")
			time.Sleep(10 * time.Second)
		}
	}

	return
}
