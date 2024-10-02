package upgrade

import (
	"io"
	"os"

	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"

	"github.com/pluralsh/plural-cli/pkg/api"
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
		Name:      "upgrade",
		Usage:     "creates an upgrade in the upgrade queue QUEUE for application REPO",
		ArgsUsage: "{queue} {repo}",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "f",
				Usage: "file containing upgrade contents, use - for stdin",
			},
		},
		Action: common.LatestVersion(common.RequireArgs(p.handleUpgrade, []string{"{queue}", "{repo}"})),
	}
}

func (p *Plural) handleUpgrade(c *cli.Context) (err error) {
	p.InitPluralClient()
	queue, repo := c.Args().Get(0), c.Args().Get(1)
	f := os.Stdin
	fname := c.String("f")
	if fname != "-" && fname != "" {
		f, err = os.Open(fname)
		if err != nil {
			return
		}
		defer f.Close()
	}

	contents, err := io.ReadAll(f)
	if err != nil {
		return
	}

	attrs, err := api.ConstructUpgradeAttributes(contents)
	if err != nil {
		return
	}

	err = p.CreateUpgrade(queue, repo, attrs)
	return api.GetErrorResponse(err, "CreateUpgrade")
}
