package up

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural-cli/pkg/bundle"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/provider"
	"github.com/pluralsh/plural-cli/pkg/utils"

	"github.com/mitchellh/go-homedir"
)

type Context struct {
	Provider       provider.Provider
	Manifest       *manifest.ProjectManifest
	Config         *config.Config
	RepoUrl        string
	StacksIdentity string
	Delims         *delims
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
		return fmt.Errorf("You never configured a repoUrl for your workspace, check `context.yaml`")
	}

	return nil
}

func Build() (*Context, error) {
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
	}, nil
}

func backfillConsoleContext(man *manifest.ProjectManifest) error {
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

	var deployKey string
	prompt := &survey.Input{
		Message: "Select a file containing a read-only deploy key for this repo (use tab to list files in the directory):",
		Default: "~/.ssh",
		Suggest: func(toComplete string) []string {
			path, err := homedir.Expand(toComplete)
			if err != nil {
				path = toComplete
			}
			files, _ := filepath.Glob(bundle.CleanPath(path) + "*")
			return files
		},
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

	console["private_key"] = contents
	ctx.Configuration["console"] = console
	return ctx.Write(path)
}
