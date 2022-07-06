package scm

import (
	"github.com/pluralsh/plural/pkg/manifest"
)

type Context struct {
	pub      string
	priv     string
	username string
	email    string
	url      string
	repoName string
}

func buildContext(context *Context) error {
	ctx := manifest.NewContext()
	ctx.Configuration = map[string]map[string]interface{}{
		"console": {
			"private_key": context.priv,
			"public_key":  context.pub,
			"passphrase":  "",
			"repo_url":    context.url,
			"git_user":    context.username,
			"git_email":   context.email,
		},
	}
	path := manifest.ContextPath()
	return ctx.Write(path)
}
