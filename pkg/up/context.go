package up

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	giturls "github.com/whilp/git-urls"
)

type Context struct {
	Provider         providerapi.Provider
	Manifest         *manifest.ProjectManifest
	Config           *config.Config
	Cloud            bool
	RepoUrl          string
	GitUsername      string
	GitPassword      string
	StacksIdentity   string
	Delims           *delims
	ImportCluster    *string
	CloudCluster     string
	dir              string
	ignorePreflights bool
}

type delims struct {
	left  string
	right string
}

func (c *Context) identifier() string {
	if c.RepoUrl == "" {
		return ""
	}

	if strings.HasPrefix(c.RepoUrl, "http") {
		parsed, err := giturls.Parse(c.RepoUrl)
		if err == nil {
			return strings.TrimSuffix(strings.TrimPrefix(parsed.Path, "/"), ".git")
		}
	}

	split := strings.Split(c.RepoUrl, ":")
	return strings.TrimSuffix(split[len(split)-1], ".git")
}

func (c *Context) changeDelims() {
	c.Delims = &delims{"[[", "]]"}
}

func (c *Context) IgnorePreflights(ignore bool) {
	c.ignorePreflights = ignore
}

func (c *Context) SetImportCluster(id string) {
	c.ImportCluster = lo.ToPtr(id)
}

func (c *Context) Backfill() error {
	context, err := manifest.FetchContext()
	if err != nil {
		return c.backfillConsoleContext(c.Manifest)
	}

	console, ok := context.Configuration["console"]
	if !ok {
		return c.backfillConsoleContext(c.Manifest)
	}

	_, hasSSH := console["private_key"]
	_, hasHTTPS := console["git_password"]
	if !hasSSH && !hasHTTPS {
		return c.backfillConsoleContext(c.Manifest)
	}

	if v, ok := console["repo_url"]; ok {
		if r, ok := v.(string); ok {
			c.RepoUrl = r
		}
	}

	if v, ok := console["git_username"]; ok {
		if s, ok := v.(string); ok {
			c.GitUsername = s
		}
	}

	if v, ok := console["git_password"]; ok {
		if s, ok := v.(string); ok {
			c.GitPassword = s
		}
	}

	if c.RepoUrl == "" {
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

func (c *Context) backfillConsoleContext(_ *manifest.ProjectManifest) error {
	path := manifest.ContextPath()
	ctx, err := manifest.FetchContext()
	if err != nil {
		ctx = manifest.NewContext()
	}

	console, ok := ctx.Configuration["console"]
	if !ok {
		console = map[string]interface{}{}
	}

	utils.Highlight("It looks like you cloned this repo before running plural up, we just need to ensure authentication is setup correctly to continue\n")

	url, err := git.GetURL()
	if err != nil {
		return err
	}

	if strings.HasPrefix(url, "http") {
		return c.backfillHTTPS(url, console, ctx, path)
	}

	return c.backfillSSH(url, console, ctx, path)
}

func (c *Context) backfillSSH(url string, console map[string]interface{}, ctx *manifest.Context, path string) error {
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

	if !c.ignorePreflights {
		if err := verifySSHKey(contents, url); err != nil {
			return fmt.Errorf("ssh key not valid for url %s, error: %w.  If you want to bypass this check, you can use the --ignore-preflights flag", url, err)
		}
	}

	console["repo_url"] = url
	console["private_key"] = contents
	ctx.Configuration["console"] = console
	c.RepoUrl = url
	return ctx.Write(path)
}

func (c *Context) backfillHTTPS(url string, console map[string]interface{}, ctx *manifest.Context, path string) error {
	utils.Highlight("If you want, you can also reclone with an SSH URL and re-run to use deploy-key authentication instead\n\n")

	var username, token string

	if err := survey.AskOne(&survey.Input{
		Message: "Enter your git username:",
		Default: "oauth2",
	}, &username, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	if err := survey.AskOne(&survey.Password{
		Message: "Enter your Personal Access Token (PAT) for this repository:",
	}, &token, survey.WithValidator(survey.Required)); err != nil {
		return err
	}

	if !c.ignorePreflights {
		if err := verifyHTTPS(username, token, url); err != nil {
			return fmt.Errorf("PAT not valid for url %s, error: %w.  If you want to bypass this check, you can use the --ignore-preflights flag", url, err)
		}
	}

	console["repo_url"] = url
	console["git_username"] = username
	console["git_password"] = token
	ctx.Configuration["console"] = console
	c.RepoUrl = url
	c.GitUsername = username
	c.GitPassword = token
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

	auth, _ := git.SSHAuth(getGitUsername(url), key, "")
	if _, err := git.Clone(auth, url, dir); err != nil {
		return err
	}
	return nil
}

func verifyHTTPS(username, password, url string) error {
	dir, err := os.MkdirTemp("", "repo")
	if err != nil {
		return err
	}
	defer func(path string) {
		_ = os.RemoveAll(path)
	}(dir)

	auth, _ := git.BasicAuth(username, password)
	if _, err := git.Clone(auth, url, dir); err != nil {
		return err
	}
	return nil
}

var (
	scpSyntax = regexp.MustCompile(`^([a-zA-Z0-9-._~]+@)?([a-zA-Z0-9._-]+):([a-zA-Z0-9./._-]+)(?:\?||$)(.*)$`)
)

func getGitUsername(url string) string {
	match := scpSyntax.FindAllStringSubmatch(url, -1)
	if len(match) > 0 {
		if match[0][1] != "" {
			return strings.TrimRight(match[0][1], "@")
		}
	}

	uname := "git"
	parsedUrl, err := giturls.Parse(url)
	if err == nil {
		uname = parsedUrl.User.Username()
	}
	return uname
}
