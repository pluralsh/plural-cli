package wkspace

import (
	"github.com/pluralsh/plural/pkg/utils"
	"os"
	"fmt"
	"time"
	"path"
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

func (w *Workspace) DestroyTerraform() error {
	repo := w.Installation.Repository
	path, err := filepath.Abs(path.Join(repo.Name, "terraform"))
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
		if err := kube.FinalizeNamespace(ns); err != nil {
			fmt.Printf("namespace finalization ignored for %s, due to %s", ns, err)
		}
	})

	os.Chdir(path)
	if err := utils.Cmd(w.Config, "terraform", "init"); err != nil {
		return err
	}
	return utils.Cmd(w.Config, "terraform", "destroy", "-auto-approve")
}
