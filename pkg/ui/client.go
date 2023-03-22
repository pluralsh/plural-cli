//go:build ui || generate

package ui

import (
	"github.com/urfave/cli"

	"github.com/pluralsh/plural/pkg/api"
)

// Client struct
type Client struct {
	ctx    *cli.Context
	client api.Client
}

func (this *Client) ListRepositories(query string) ([]*api.Repository, error) {
	return this.client.ListRepositories(query)
}

func (this *Client) ListRecipes(repo string, provider string) ([]*api.Recipe, error) {
	return this.client.ListRecipes(repo, provider)
}

// NewClient creates a new proxy client struct
func NewClient(client api.Client, ctx *cli.Context) *Client {
	return &Client{
		ctx:    ctx,
		client: client,
	}
}
