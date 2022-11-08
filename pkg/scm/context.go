package scm

import (
	"github.com/pluralsh/plural/pkg/manifest"
)

type GitProvider string

const (
	GitHub GitProvider = "github"
	GitLab GitProvider = "gitlab"
)

type Context struct {
	pub         string
	priv        string
	username    string
	email       string
	url         string
	repoName    string
	gitProvider GitProvider
	token       string
}

func buildContext(context *Context) error {
	ctx := manifest.NewContext()
	ctx.Configuration = map[string]map[string]interface{}{
		"console": {
			"private_key":  context.priv,
			"public_key":   context.pub,
			"passphrase":   "",
			"repo_url":     context.url,
			"git_user":     context.username,
			"git_email":    context.email,
			"git_provider": context.gitProvider,
			"token":        context.token,
		},
	}
	path := manifest.ContextPath()
	return ctx.Write(path)
}
