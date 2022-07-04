package wkspace

import (
	"fmt"
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
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
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

	projPath, _ := filepath.Abs("workspace.yaml")
	project, err := manifest.ReadProject(projPath)
	if err != nil {
		return nil, err
	}

	prov, err := provider.FromManifest(project)
	if err != nil {
		return nil, err
	}

	conf := config.Read()
	ctx, err := manifest.ReadContext(manifest.ContextPath())
	if err != nil {
		return nil, err
	}

	manifestPath := manifestPath(inst.Repository.Name)
	man, err := manifest.Read(manifestPath)
	var links *manifest.Links
	if err == nil {
		links = man.Links
	}

	wk := &Workspace{
		Provider:     prov,
		Installation: inst,
		Charts:       ci,
		Terraform:    ti,
		Config:       &conf,
		Context:      ctx,
		Manifest:     project,
		Links:        links,
	}
	return wk, nil
}

func Configured(repo string) bool {
	ctx, err := manifest.ReadContext(manifest.ContextPath())
	if err != nil {
		return false
	}

	_, ok := ctx.Configuration[repo]
	return ok
}

func (wk *Workspace) PrintLinks() {
	if wk.Links == nil {
		return
	}

	fmt.Printf("\n")
	doPrintLinks("helm", wk.Links.Helm)
	doPrintLinks("terraform", wk.Links.Terraform)
}

func doPrintLinks(name string, links map[string]string) {
	if len(links) == 0 {
		return
	}

	utils.Highlight("configured %s links:\n", name)
	for name, path := range links {
		fmt.Printf("\t%s ==> %s\n", name, path)
	}

	fmt.Printf("\n")
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
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	if err := mkdir(pathing.SanitizeFilepath(filepath.Join(repoRoot, repo.Name))); err != nil {
		return err
	}

	path, _ := manifest.ManifestPath(repo.Name)
	prev, err := manifest.Read(path)
	if err != nil {
		prev = &manifest.Manifest{}
	}

	man := wk.BuildManifest(prev)
	if err := man.Write(path); err != nil {
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

func (wk *Workspace) requiresWait() bool {
	for _, ci := range wk.Charts {
		if ci.Version.Dependencies.Wait {
			return true
		}
	}

	for _, ti := range wk.Terraform {
		if ti.Version.Dependencies.Wait {
			return true
		}
	}
	return false
}

func (wk *Workspace) buildExecution(repoRoot string) error {
	name := wk.Installation.Repository.Name
	wkspaceRoot := filepath.Join(repoRoot, name)

	if err := mkdir(pathing.SanitizeFilepath(filepath.Join(wkspaceRoot, ".plural"))); err != nil {
		return err
	}

	onceFile := pathing.SanitizeFilepath(filepath.Join(wkspaceRoot, ".plural", "ONCE"))
	if err := ioutil.WriteFile(onceFile, []byte("once"), 0644); err != nil {
		return err
	}

	nonceFile := pathing.SanitizeFilepath(filepath.Join(wkspaceRoot, ".plural", "NONCE"))
	if err := ioutil.WriteFile(nonceFile, []byte(crypto.RandString(32)), 0644); err != nil {
		return err
	}

	if err := executor.Ignore(wkspaceRoot); err != nil {
		return err
	}

	exec, _ := executor.GetExecution(pathing.SanitizeFilepath(filepath.Join(wkspaceRoot)), "deploy")

	return executor.DefaultExecution(name, exec).Flush(repoRoot)
}

func (wk *Workspace) buildDiff(repoRoot string) error {
	name := wk.Installation.Repository.Name
	wkspaceRoot := pathing.SanitizeFilepath(filepath.Join(repoRoot, name))

	d, _ := diff.GetDiff(pathing.SanitizeFilepath(filepath.Join(wkspaceRoot)), "diff")

	return diff.DefaultDiff(name, d).Flush(repoRoot)
}

func DiffedRepos() ([]string, error) {
	files, err := git.Modified()
	repos := make(map[string]bool)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		// we don't want to respect the OS separators here, it is always a forwards slash on git
		parts := strings.Split(file, "/")
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
	for repo := range repos {
		result[count] = repo
		count++
	}

	return result, nil
}

func isRepo(name string) bool {
	repoRoot, err := git.Root()
	if err != nil {
		return false
	}

	return utils.Exists(pathing.SanitizeFilepath(filepath.Join(repoRoot, name, "manifest.yaml")))
}

func mkdir(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func manifestPath(repo string) string {
	path, _ := filepath.Abs(pathing.SanitizeFilepath(filepath.Join(repo, "manifest.yaml")))
	return path
}
