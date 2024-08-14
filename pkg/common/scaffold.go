package common

import (
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/scaffold"
	"github.com/urfave/cli"
)

func HandleScaffold(c *cli.Context) error {
	client := api.NewClient()
	return scaffold.ApplicationScaffold(client)
}
