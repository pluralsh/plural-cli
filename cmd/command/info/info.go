package info

import (
	"fmt"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/scaffold"
	"github.com/urfave/cli"
)

type Plural struct {
	client.Plural
}

func Command(clients client.Plural) cli.Command {
	p := Plural{
		Plural: clients,
	}
	return cli.Command{
		Name:      "info",
		Usage:     "Get information for your installation of APP",
		ArgsUsage: "{app}",
		Action:    common.LatestVersion(common.RequireArgs(common.Owned(common.Rooted(p.info)), []string{"{app}"})),
	}
}
func (p *Plural) info(c *cli.Context) error {
	p.InitPluralClient()
	repo := c.Args().Get(0)
	installation, err := p.GetInstallation(repo)
	if err != nil {
		return api.GetErrorResponse(err, "GetInstallation")
	}
	if installation == nil {
		return fmt.Errorf("You have not installed %s", repo)
	}

	return scaffold.Notes(installation)
}
