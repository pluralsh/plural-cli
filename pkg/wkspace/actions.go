package wkspace

import (
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/executor"
	"os"
	"fmt"
	"time"
	"strings"
	"path/filepath"
)

func execSuppressed(command string, args ...string) (err error) {
	for retry := 2; retry >= 0; retry-- {
		utils.Highlight("%s %s ~> ", command, strings.Join(args, " "))
		cmd, out := executor.SuppressedCommand(command, args...)
		err = executor.RunCommand(cmd, out)
		if err == nil {
			break
		}
		fmt.Printf("retrying command, number of retries remaining: %d\n", retry)
	}

	return
}

func (w *Workspace) DestroyHelm() error {
	// ensure current kubeconfig is correct before destroying stuff
	w.Provider.KubeConfig()
	name := w.Installation.Repository.Name

	ns := w.Config.Namespace(name)
	if err := execSuppressed("helm", "get", "values", name, "-n", ns); err != nil {
		fmt.Println("Helm already uninstalled, continuing...\n")
		return nil
	}

	return execSuppressed("helm", "del", name, "-n", ns)
}

func (w *Workspace) Bounce() error {
	return w.ToMinimal().BounceHelm()
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
	repoRoot, err := utils.RepoRoot()
	if err != nil {
		return err
	}

	deployfile := filepath.Join(repoRoot, repo.Name, "deploy.hcl")
	os.Remove(deployfile)
	return nil
}

func (w *Workspace) DestroyTerraform() error {
	repo := w.Installation.Repository
	path, err := filepath.Abs(filepath.Join(repo.Name, "terraform"))
	if err != nil {
		return err
	}

	time.AfterFunc(1 * time.Minute, func() {
		kube, err := utils.Kubernetes()
		if err != nil {
			fmt.Println("could not set up k8s client due to %s", err)
			return
		}

		ns := w.Config.Namespace(repo.Name)
		kube.FinalizeNamespace(ns)
	})

	os.Chdir(path)
	if err := execSuppressed("terraform", "init"); err != nil {
		return err
	}

	return execSuppressed("terraform", "destroy", "-auto-approve")
}
