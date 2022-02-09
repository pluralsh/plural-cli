package main

import (
	"github.com/urfave/cli"
	"github.com/pluralsh/plural/pkg/scaffold"
	"github.com/pluralsh/plural/pkg/api"
)

func handleScaffold(c *cli.Context) error {
	client := api.NewClient()
	return scaffold.ApplicationScaffold(client)
}