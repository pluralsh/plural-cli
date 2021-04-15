package wkspace

import (
	"github.com/pluralsh/plural/pkg/utils"
	"os"
	"path"
	"path/filepath"
)

func (w *Workspace) DestroyHelm() error {
	// ensure current kubeconfig is correct before destroying stuff
	w.Provider.KubeConfig()
	name := w.Installation.Repository.Name
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

	os.Chdir(path)
	return utils.Cmd(w.Config, "terraform", "destroy", "-auto-approve")
}
