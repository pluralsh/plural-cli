package up

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/pluralsh/plural-cli/pkg/kubernetes"
	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
		{dir: "./terraform/mgmt", cmd: "init", args: []string{"-upgrade"}},
		{dir: "./terraform/mgmt", cmd: "apply", args: []string{"-auto-approve"}, retries: 1},
	}); err != nil {
		return err
	}

	if ctx.ImportCluster != nil {
		prov := ctx.Provider.Name()
		if err := ctx.templateFrom(ctx.path(fmt.Sprintf("templates/setup/mgmt/%s.tf", prov)), "terraform/mgmt/plural.tf"); err != nil {
			return err
		}

		if err := runAll([]terraformCmd{
			{dir: "./terraform/mgmt", cmd: "init", args: []string{"-upgrade"}},
			{dir: "./terraform/mgmt", cmd: "import", args: []string{"plural_cluster.mgmt", *ctx.ImportCluster}},
			{dir: "./terraform/mgmt", cmd: "apply", args: []string{"-auto-approve"}, retries: 1},
		}); err != nil {
			return err
		}
	}

	stateCmd := &terraformCmd{dir: "./terraform/mgmt"}
	outs, err := stateCmd.outputs()
	if err != nil {
		return err
	}

	ctx.StacksIdentity = stacksRole(outs)

	if err := ctx.afterSetup(); err != nil {
		return err
	}

	if !ctx.Cloud {
		subdomain := ctx.Manifest.Network.Subdomain
		if err := testDns(fmt.Sprintf("console.%s", subdomain)); err != nil {
			return err
		}

		if err := ping(fmt.Sprintf("https://console.%s", subdomain)); err != nil {
			return err
		}
	}

	if err := commit(); err != nil {
		return err
	}

	utils.Highlight("\nSetting up gitops management...\n")

	if err := runAll([]terraformCmd{
		{dir: "./terraform/apps", cmd: "init", args: []string{"-upgrade"}},
		{dir: "./terraform/apps", cmd: "apply", args: []string{"-auto-approve"}, retries: 1},
	}); err != nil {
		return err
	}

	return ctx.Prune()
}

func (ctx *Context) Destroy() error {
	if err := ctx.DestroyNamespace("plural-runtime"); err != nil {
		return err
	}
	return runAll([]terraformCmd{
		{dir: "./terraform/mgmt", cmd: "init", args: []string{"-upgrade"}},
		{dir: "./terraform/mgmt", cmd: "destroy", args: []string{"-auto-approve"}, retries: 2},
	})
}

func (ctx *Context) DestroyNamespace(name string) error {
	utils.Highlight("\nCleaning up namespace %s...\n", name)
	// ensure current kubeconfig is correct before destroying stuff
	if err := ctx.Provider.KubeConfig(); err != nil {
		return err
	}
	kube, err := kubernetes.Kubernetes()
	if err != nil {
		utils.Error("Could not set up k8s client due to %s\n", err)
		return err
	}
	c := context.Background()
	namespace, err := kube.GetClient().CoreV1().Namespaces().Get(c, name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if namespace != nil {
		return kube.GetClient().CoreV1().Namespaces().Delete(c, name, metav1.DeleteOptions{
			GracePeriodSeconds: lo.ToPtr(int64(0)),
		})
	}

	return nil

}

func runAll(cmds []terraformCmd) error {
	for _, cmd := range cmds {
		if err := cmd.run(); err != nil {
			return err
		}
	}

	return nil
}

func (tf *terraformCmd) outputs() (map[string]Output, error) {
	outputs := map[string]Output{}
	cmd := exec.Command("terraform", "output", "-json")
	cmd.Dir = tf.dir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(out, &outputs); err != nil {
		return nil, err
	}

	return outputs, nil
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
