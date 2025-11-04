package up

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/samber/lo"

	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/provider"
	providerapi "github.com/pluralsh/plural-cli/pkg/provider/api"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/git"

	"github.com/mitchellh/go-homedir"
)

type Context struct {
	Provider       providerapi.Provider
	Manifest       *manifest.ProjectManifest
	Config         *config.Config
	Cloud          bool
	RepoUrl        string
	StacksIdentity string
	Delims         *delims
	ImportCluster  *string
	CloudCluster   string
	dir            string
}

type delims struct {
	left  string
	right string
}

func (ctx *Context) identifier() string {
	if ctx.RepoUrl == "" {
		return ""
	}

	split := strings.Split(ctx.RepoUrl, ":")
	return strings.TrimSuffix(split[len(split)-1], ".git")
}

func (ctx *Context) changeDelims() {
	ctx.Delims = &delims{"[[", "]]"}
}

func (ctx *Context) SetImportCluster(id string) {
	ctx.ImportCluster = lo.ToPtr(id)
}

func (ctx *Context) Backfill() error {
	context, err := manifest.FetchContext()
	if err != nil {
		return backfillConsoleContext(ctx.Manifest)
	}

	console, ok := context.Configuration["console"]
	if !ok {
		return backfillConsoleContext(ctx.Manifest)
	}

	if _, ok = console["private_key"]; !ok {
		return backfillConsoleContext(ctx.Manifest)
	}

	if v, ok := console["repo_url"]; ok {
		if r, ok := v.(string); ok {
			ctx.RepoUrl = r
		}
	}

	if ctx.RepoUrl == "" {
		return fmt.Errorf("you never configured a repoUrl for your workspace, check `context.yaml`")
	}

	return nil
}

func Build(cloud bool) (*Context, error) {
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
	return &Context{
		Provider: prov,
		Config:   &conf,
		Manifest: project,
		Cloud:    cloud,
	}, nil
}

func backfillConsoleContext(_ *manifest.ProjectManifest) error {
	path := manifest.ContextPath()
	ctx, err := manifest.FetchContext()
	if err != nil {
		ctx = manifest.NewContext()
	}

	console, ok := ctx.Configuration["console"]
	if !ok {
		console = map[string]interface{}{}
	}

	utils.Highlight("It looks like you cloned this repo before running plural up, we just need you to generate and give us a deploy key to continue\n")
	utils.Highlight("If you want, you can use `plural crypto ssh-keygen` to generate a keypair to use as a deploy key as well\n\n")

	files, err := filepath.Glob(filepath.Join(os.Getenv("HOME"), ".ssh", "*"))
	if err != nil {
		return err
	}

	var deployKey string
	prompt := &survey.Select{
		Message: "Select a file containing a read-only deploy key for this repo (use tab to list files in the directory):",
		Options: files,
	}

	opts := []survey.AskOpt{survey.WithValidator(survey.Required)}
	if err := survey.AskOne(prompt, &deployKey, opts...); err != nil {
		return err
	}

	keyPath, err := homedir.Expand(deployKey)
	if err != nil {
		return err
	}

	contents, err := utils.ReadFile(keyPath)
	if err != nil {
		return err
	}

	url, err := git.GetURL()
	if err != nil {
		return err
	}

	if strings.HasPrefix(url, "http") {
		return fmt.Errorf("found non-ssh upstream url %s, please reclone the repo with SSH and retry", url)
	}

	if err := verifySSHKey(contents, url); err != nil {
		return fmt.Errorf("ssh key not valid for url %s, error: %w", url, err)
	}

	console["repo_url"] = url
	console["private_key"] = contents
	ctx.Configuration["console"] = console
	return ctx.Write(path)
}

func verifySSHKey(key, url string) error {
	dir, err := os.MkdirTemp("", "repo")
	if err != nil {
		return err
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			return
		}
	}(dir)
	auth, _ := git.SSHAuth("git", key, "")
	if _, err := git.Clone(auth, url, dir); err != nil {
		return err
	}
	return nil
}
