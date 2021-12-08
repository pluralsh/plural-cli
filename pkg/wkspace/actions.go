package wkspace

import (
	"github.com/pluralsh/plural/pkg/utils"
	"os"
	"fmt"
	"time"
	"path/filepath"
)

func (w *Workspace) DestroyHelm() error {
	// ensure current kubeconfig is correct before destroying stuff
	w.Provider.KubeConfig()
	name := w.Installation.Repository.Name

	err := utils.Cmd(w.Config, "helm", "get", "values", name, "-n", w.Config.Namespace(name))
	if err != nil {
		fmt.Println("Helm already uninstalled, continuing...")
		return nil
	}

	return utils.Cmd(w.Config, "helm", "del", name, "-n", w.Config.Namespace(name))
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
	if err := utils.Cmd(w.Config, "terraform", "init"); err != nil {
		return err
	}

	return utils.Cmd(w.Config, "terraform", "destroy", "-auto-approve")
}
