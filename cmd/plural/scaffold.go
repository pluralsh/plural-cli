package plural

import (
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/scaffold"
	"github.com/urfave/cli"
)

func handleScaffold(c *cli.Context) error {
	client := api.NewClient()
	return scaffold.ApplicationScaffold(client)
}
