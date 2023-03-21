//go:build ui || generate

package ui

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/pluralsh/plural/pkg/api"
)

// Client struct
type Client struct {
	ctx    *cli.Context
	client api.Client
}

func (this *Client) ListRepositories() {
	repos, err := this.client.ListRepositories(this.ctx.String("query"))
	if err != nil {
		return
	}

	for _, repo := range repos {
		fmt.Println(repo.Id)
		fmt.Println(repo.Name)
		fmt.Println()
	}
}

// NewClient creates a new proxy client struct
func NewClient(client api.Client, ctx *cli.Context) *Client {
	return &Client{
		ctx:    ctx,
		client: client,
	}
}
