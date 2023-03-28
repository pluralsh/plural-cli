//go:build ui || generate

package ui

import (
	"github.com/urfave/cli"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/manifest"
)

// Client struct
type Client struct {
	ctx    *cli.Context
	client api.Client
}

func (this *Client) Token() string {
	conf := config.Read()

	return conf.Token
}

func (this *Client) Project() *manifest.ProjectManifest {
	project, err := manifest.FetchProject()
	if err != nil {
		return nil
	}

	return project
}

func (this *Client) Context() *manifest.Context {
	context, err := manifest.FetchContext()
	if err != nil {
		return nil
	}

	return context
}

// NewClient creates a new proxy client struct
func NewClient(client api.Client, ctx *cli.Context) *Client {
	return &Client{
		ctx:    ctx,
		client: client,
	}
}
