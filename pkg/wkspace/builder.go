package wkspace

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/michaeljguarino/forge/pkg/api"
	"github.com/michaeljguarino/forge/pkg/config"
	"github.com/michaeljguarino/forge/pkg/crypto"
	"github.com/michaeljguarino/forge/pkg/diff"
	"github.com/michaeljguarino/forge/pkg/executor"
	"github.com/michaeljguarino/forge/pkg/manifest"
	"github.com/michaeljguarino/forge/pkg/provider"
	"github.com/michaeljguarino/forge/pkg/utils"
)

type Workspace struct {
	Provider     provider.Provider
	Installation *api.Installation
	Charts       []api.ChartInstallation
	Terraform    []api.TerraformInstallation
	Config       *config.Config
	Manifest     *manifest.ProjectManifest
}

func New(client *api.Client, inst *api.Installation) (*Workspace, error) {
	ci, ti, err := client.GetPackageInstallations(inst.Repository.Id)
	if err != nil {
		return nil, err
	}

	projPath, _ := filepath.Abs("workspace.yaml")
	project, err := manifest.ReadProject(projPath)
	if err != nil {
		return nil, err
	}

	var prov provider.Provider
	manifestPath := manifestPath(&inst.Repository)
	if utils.Exists(manifestPath) {
		man, err := manifest.Read(manifestPath)
		if err != nil {
			return nil, err
		} 

		prov, err = provider.FromManifest(man)
		if err != nil {
			return nil, err
		}
	} else {
		prov, err = provider.Select()
		if err != nil {
			return nil, err
		}
	}

	conf := config.Read()
	wk := &Workspace{
		Provider: prov,
		Installation: inst,
		Charts: ci,
		Terraform: ti,
		Config: &conf,
		Manifest: project,
	}
	return wk, nil
}

func (wk *Workspace) ToMinimal() *MinimalWorkspace {
	return &MinimalWorkspace{
		Name:     wk.Installation.Repository.Name,
		Provider: wk.Provider,
		Config:   wk.Config,
		Manifest: wk.Manifest,
	}
}

func (wk *Workspace) Prepare() error {
	repo := wk.Installation.Repository
	repoRoot, err := utils.RepoRoot()
	if err != nil {
		return err
	}

	manifest := wk.BuildManifest()
	if err := mkdir(filepath.Join(repoRoot, repo.Name)); err != nil {
		return err
	}

	if err := manifest.Write(manifestPath(&repo)); err != nil {
		return err
	}

	if err := wk.buildExecution(repoRoot); err != nil {
		return err
	}

	if err := wk.buildDiff(repoRoot); err != nil {
		return err
	}

	return nil
}

func (wk *Workspace) buildExecution(repoRoot string) error {
	name := wk.Installation.Repository.Name
	wkspaceRoot := filepath.Join(repoRoot, name)

	if err := mkdir(filepath.Join(wkspaceRoot, ".forge")); err != nil {
		return err
	}

	onceFile := filepath.Join(wkspaceRoot, ".forge", "ONCE")
	if err := ioutil.WriteFile(onceFile, []byte("once"), 0644); err != nil {
		return err
	}

	nonceFile := filepath.Join(wkspaceRoot, ".forge", "NONCE")
	if err := ioutil.WriteFile(nonceFile, []byte(crypto.RandString(32)), 0644); err != nil {
		return err
	}

	if err := executor.Ignore(wkspaceRoot); err != nil {
		return err
	}

	exec, _ := executor.GetExecution(filepath.Join(wkspaceRoot), "deploy")

	return executor.DefaultExecution(name, exec).Flush(repoRoot)
}

func (wk *Workspace) buildDiff(repoRoot string) error {
	name := wk.Installation.Repository.Name
	wkspaceRoot := filepath.Join(repoRoot, name)

	d, _ := diff.GetDiff(filepath.Join(wkspaceRoot), "diff")

	return diff.DefaultDiff(name, d).Flush(repoRoot)
}

func mkdir(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func manifestPath(repo *api.Repository) string {
	path, _ := filepath.Abs(filepath.Join(repo.Name, "manifest.yaml"))
	return path
}
