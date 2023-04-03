package wkspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pluralsh/plural/pkg/executor"
	"github.com/pluralsh/plural/pkg/helm"
	"github.com/pluralsh/plural/pkg/kubernetes"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"go.mercari.io/hcledit"
	"helm.sh/helm/v3/pkg/action"
)

type checker func(s string) bool

var alwaysErr checker = func(s string) bool { return false }

func (c checker) execSuppressed(command string, args ...string) (err error) {
	for retry := 2; retry >= 0; retry-- {
		utils.Highlight("%s %s ~> ", command, strings.Join(args, " "))
		cmd, out := executor.SuppressedCommand(command, args...)
		err = executor.RunCommand(cmd, out)
		if err == nil || c(out.Format()) {
			break
		}
		fmt.Printf("retrying command, number of retries remaining: %d\n", retry)
	}

	return
}

func (w *Workspace) DestroyHelm() error {
	// ensure current kubeconfig is correct before destroying stuff
	if err := w.Provider.KubeConfig(); err != nil {
		return err
	}

	name := w.Installation.Repository.Name
	namespace := w.Config.Namespace(name)
	var err error
	for retry := 2; retry >= 0; retry-- {
		err = uninstallHelm(name, namespace)
		if err == nil {
			break
		}
		fmt.Printf("retrying command, number of retries remaining: %d\n", retry)
	}

	return err

}

func (w *Workspace) Bounce() error {
	return w.ToMinimal().BounceHelm(false)
}

func (w *Workspace) HelmDiff() error {
	return w.ToMinimal().DiffHelm()
}

func (w *Workspace) Destroy() error {
	if err := w.DestroyHelm(); err != nil {
		return err
	}

	if err := w.DestroyTerraform(); err != nil {
		return err
	}

	return w.Reset()
}

func (w *Workspace) Reset() error {
	repo := w.Installation.Repository
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	deployFile := pathing.SanitizeFilepath(filepath.Join(repoRoot, repo.Name, "deploy.hcl"))
	editor, err := hcledit.ReadFile(deployFile)
	if err != nil {
		return err
	}
	if err := editor.Update("step.*.sha", ""); err != nil {
		return err
	}
	if err := editor.OverWriteFile(); err != nil {
		return err
	}

	return nil
}

func (w *Workspace) DestroyTerraform() error {
	repo := w.Installation.Repository
	path, err := filepath.Abs(pathing.SanitizeFilepath(filepath.Join(repo.Name, "terraform")))
	if err != nil {
		return err
	}

	time.AfterFunc(1*time.Minute, func() {
		kube, err := kubernetes.Kubernetes()
		if err != nil {
			fmt.Printf("Could not set up k8s client due to %s\n", err)
			return
		}

		ns := w.Config.Namespace(repo.Name)
		if err := kube.FinalizeNamespace(ns); err != nil {
			return
		}
	})

	if err := os.Chdir(path); err != nil {
		return err
	}
	if err := alwaysErr.execSuppressed("terraform", "init", "-upgrade"); err != nil {
		return err
	}

	return alwaysErr.execSuppressed("terraform", "destroy", "-auto-approve")
}

func uninstallHelm(name, namespace string) error {
	exists, err := isReleaseAvailable(name, namespace)
	if err != nil {
		return err
	}
	if exists {
		actionConfig, err := helm.GetActionConfig(namespace)
		if err != nil {
			return err
		}
		client := action.NewUninstall(actionConfig)

		_, err = client.Run(name)
		if err != nil {
			return err
		}
	}
	return nil
}

func isReleaseAvailable(name, namespace string) (bool, error) {
	actionConfig, err := helm.GetActionConfig(namespace)
	if err != nil {
		return false, err
	}
	client := action.NewList(actionConfig)
	resp, err := client.Run()
	if err != nil {
		return false, err
	}
	for _, rel := range resp {
		if rel.Name == name {
			return true, nil
		}
	}
	return false, nil
}
