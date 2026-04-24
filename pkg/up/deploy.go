package up

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/kubernetes"

	"github.com/pluralsh/plural-cli/pkg/utils"
)

type terraformCmd struct {
	dir     string
	cmd     string
	args    []string
	retries int
}

var (
	checkpoints = []string{
		"init",
		"import",
		"apply:import",
		"commit",
		"apps",
		"prune:cloud",
		"prune:mgmt",
	}

	priorities = map[string]int{}
)

func init() {
	for i, c := range checkpoints {
		priorities[c] = i
	}
}

func (c *Context) runCheckpoint(current, checkpoint string, fn func() error) error {
	if current == "" || priorities[checkpoint] > priorities[current] {
		err := fn()
		if err == nil {
			c.Manifest.Checkpoint = checkpoint
		}
		return err
	}

	utils.Highlight("Skipping checkpoint %s, ran up to %s previously\n", checkpoint, current)

	return nil
}

func (c *Context) Deploy(commit func() error) error {
	if c.Provider.Name() == api.BYOK && c.Cloud {
		return nil
	}

	if c.Provider.Name() == api.BYOK {
		return c.deployBYOK(commit)
	}

	if err := c.Provider.CreateBucket(); err != nil {
		return err
	}
	defer c.Manifest.Flush()

	if err := c.runCheckpoint(c.Manifest.Checkpoint, "init", func() error {
		return runAll([]terraformCmd{
			{dir: "./terraform/mgmt", cmd: "init", args: []string{"-upgrade"}},
			{dir: "./terraform/mgmt", cmd: "apply", args: []string{"-auto-approve"}, retries: 1},
		})
	}); err != nil {
		return err
	}

	if c.ImportCluster != nil {
		prov := c.Provider.Name()
		if err := c.templateFrom(c.path(fmt.Sprintf("templates/setup/mgmt/%s.tf", prov)), "terraform/mgmt/plural.tf"); err != nil {
			return err
		}

		if err := c.runCheckpoint(c.Manifest.Checkpoint, "import", func() error {
			return runAll([]terraformCmd{
				{dir: "./terraform/mgmt", cmd: "init", args: []string{"-upgrade"}},
				{dir: "./terraform/mgmt", cmd: "import", args: []string{"plural_cluster.mgmt", *c.ImportCluster}},
			})
		}); err != nil {
			return err
		}

		if err := c.runCheckpoint(c.Manifest.Checkpoint, "apply:import", func() error {
			return runAll([]terraformCmd{
				{dir: "./terraform/mgmt", cmd: "apply", args: []string{"-auto-approve"}, retries: 1},
			})
		}); err != nil {
			return err
		}
	}

	stateCmd := &terraformCmd{dir: "./terraform/mgmt"}
	outs, err := stateCmd.outputs()
	if err != nil {
		return err
	}

	c.StacksIdentity = stacksRole(outs)

	if err := c.afterSetup(); err != nil {
		return err
	}

	if !c.Cloud {
		subdomain := c.Manifest.Network.Subdomain
		if err := testDns(fmt.Sprintf("console.%s", subdomain)); err != nil {
			return err
		}

		if err := ping(fmt.Sprintf("https://console.%s", subdomain)); err != nil {
			return err
		}
	}

	if err := c.runCheckpoint(c.Manifest.Checkpoint, "commit", func() error {
		utils.Highlight("\nSetting up gitops management, first lets commit the changes made up to this point...\n\n")
		return commit()
	}); err != nil {
		return err
	}

	if err := c.runCheckpoint(c.Manifest.Checkpoint, "apps", func() error {
		return runAll([]terraformCmd{
			{dir: "./terraform/mgmt", cmd: "apply", args: []string{"-auto-approve"}},
			{dir: "./terraform/apps", cmd: "init", args: []string{"-upgrade"}},
			{dir: "./terraform/apps", cmd: "apply", args: []string{"-auto-approve"}, retries: 1},
		})
	}); err != nil {
		return err
	}

	return c.Prune()
}

func (c *Context) Destroy() error {
	utils.Highlight("Destroying management cluster terraform stack in terraform/mgmt...\n\n")
	if c.Cloud {
		return runAll([]terraformCmd{
			{dir: "./terraform/mgmt", cmd: "init", args: []string{"-upgrade"}},
			{dir: "./terraform/mgmt", cmd: "state", args: []string{"rm", "plural_cluster.mgmt"}},
			{dir: "./terraform/mgmt", cmd: "destroy", args: []string{"-auto-approve"}, retries: 2},
		})
	}

	return runAll([]terraformCmd{
		{dir: "./terraform/mgmt", cmd: "init", args: []string{"-upgrade"}},
		{dir: "./terraform/mgmt", cmd: "destroy", args: []string{"-auto-approve"}, retries: 2},
	})
}

func (c *Context) DestroyNamespace(name string) error {
	utils.Highlight("\nCleaning up namespace %s...\n", name)
	// ensure current kubeconfig is correct before destroying stuff
	if err := c.Provider.KubeConfig(); err != nil {
		return err
	}
	kube, err := kubernetes.Kubernetes()
	if err != nil {
		utils.Error("Could not set up k8s client due to %s\n", err)
		return err
	}
	ctx := context.Background()
	namespace, err := kube.GetClient().CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if namespace != nil {
		return kube.GetClient().CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{
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

// deployBYOK runs only the terraform/mgmt apply which installs the Plural console
// onto an already-existing (BYOK) local Kubernetes cluster. No cloud infrastructure
// or apps terraform is executed.
func (c *Context) deployBYOK(commit func() error) error {
	defer c.Manifest.Flush()

	if err := c.runCheckpoint(c.Manifest.Checkpoint, "init", func() error {
		return runAll([]terraformCmd{
			{dir: "./terraform/mgmt", cmd: "init", args: []string{"-upgrade"}},
			{dir: "./terraform/mgmt", cmd: "apply", args: []string{"-auto-approve"}, retries: 1},
		})
	}); err != nil {
		return err
	}

	if err := c.runCheckpoint(c.Manifest.Checkpoint, "commit", func() error {
		utils.Highlight("\nCommitting generated gitops configuration...\n\n")
		return commit()
	}); err != nil {
		return err
	}

	subdomain := c.Manifest.Network.Subdomain
	if err := waitForConsole(); err != nil {
		return err
	}

	utils.Success("Console is up! Access it at https://console.%s\n", subdomain)

	if err := c.afterSetup(); err != nil {
		return err
	}

	if err := c.runCheckpoint(c.Manifest.Checkpoint, "apps", func() error {
		return runAll([]terraformCmd{
			{dir: "./terraform/mgmt", cmd: "apply", args: []string{"-auto-approve"}},
			{dir: "./terraform/apps", cmd: "init", args: []string{"-upgrade"}},
			{dir: "./terraform/apps", cmd: "apply", args: []string{"-auto-approve"}, retries: 1},
		})
	}); err != nil {
		return err
	}

	return c.pruneBYOK()
}
