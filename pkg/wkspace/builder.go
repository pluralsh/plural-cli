package wkspace

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/crypto"
	"github.com/pluralsh/plural/pkg/diff"
	"github.com/pluralsh/plural/pkg/executor"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
)

type Workspace struct {
	Provider     provider.Provider
	Installation *api.Installation
	Charts       []*api.ChartInstallation
	Terraform    []*api.TerraformInstallation
	Config       *config.Config
	Manifest     *manifest.ProjectManifest
	Context      *manifest.Context
	Links        *manifest.Links
}

func New(client *api.Client, inst *api.Installation) (*Workspace, error) {
	ci, ti, err := client.GetPackageInstallations(inst.Repository.Id)
	if err != nil {
		return nil, err
	}

	manifestPath := manifestPath(inst.Repository.Name)
	prov, err := provider.Bootstrap(manifestPath, true)
	if err != nil {
		return nil, err
	}

	projPath, _ := filepath.Abs("workspace.yaml")
	project, err := manifest.ReadProject(projPath)
	if err != nil {
		return nil, err
	}

	conf := config.Read()
	ctx, err := manifest.ReadContext(manifest.ContextPath())
	if err != nil {
		return nil, err
	}

	man, err := manifest.Read(manifestPath)
	var links *manifest.Links
	if err == nil {
		links = man.Links
	}

	wk := &Workspace{
		Provider: prov,
		Installation: inst,
		Charts: ci,
		Terraform: ti,
		Config: &conf,
		Context: ctx,
		Manifest: project,
		Links: links,
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

	path, _ := manifest.ManifestPath(repo.Name)
	prev, err := manifest.Read(path)
	if err != nil {
		prev = &manifest.Manifest{}
	}

	manifest := wk.BuildManifest(prev)
	if err := mkdir(filepath.Join(repoRoot, repo.Name)); err != nil {
		return err
	}

	if err := manifest.Write(path); err != nil {
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

	if err := mkdir(filepath.Join(wkspaceRoot, ".plural")); err != nil {
		return err
	}

	onceFile := filepath.Join(wkspaceRoot, ".plural", "ONCE")
	if err := ioutil.WriteFile(onceFile, []byte("once"), 0644); err != nil {
		return err
	}

	nonceFile := filepath.Join(wkspaceRoot, ".plural", "NONCE")
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

func DiffedRepos() ([]string, error) {
	files, err := utils.ChangedFiles()
	repos := make(map[string]bool)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		parts := strings.Split(file, string([]byte{ filepath.Separator }))
		if len(parts) <= 1 {
			continue
		}

		maybeRepo := parts[0]
		if utils.Exists(manifestPath(maybeRepo)) && file != manifestPath(maybeRepo) {
			repos[maybeRepo] = true
		}
	}

	result := make([]string, len(repos))
	count := 0
	for repo, _ := range repos {
		result[count] = repo
		count++
	}
	return result, nil
}

func mkdir(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func manifestPath(repo string) string {
	path, _ := filepath.Abs(filepath.Join(repo, "manifest.yaml"))
	return path
}
