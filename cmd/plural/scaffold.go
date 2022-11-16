package main

import (
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/scaffold"
	"github.com/urfave/cli/v2"
)

func handleScaffold(c *cli.Context) error {
	client := api.NewClient()
	return scaffold.ApplicationScaffold(client)
}
