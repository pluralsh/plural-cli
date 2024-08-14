package buildcmd

import (
	"fmt"

	"github.com/pluralsh/plural-cli/cmd/command/crypto"
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/errors"
	"github.com/urfave/cli"
)

type Plural struct {
	Plural client.Plural
}

func Command(clients client.Plural) cli.Command {
	p := Plural{
		Plural: clients,
	}
	return cli.Command{
		Name:    "build",
		Aliases: []string{"bld"},
		Usage:   "builds your workspace",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "only",
				Usage: "repository to (re)build",
			},
			cli.BoolFlag{
				Name:  "force",
				Usage: "force workspace to build even if remote is out of sync",
			},
		},
		Action: common.Tracked(common.Rooted(common.LatestVersion(common.Owned(common.UpstreamSynced(p.build)))), "cli.build"),
	}
}

func (p *Plural) build(c *cli.Context) error {
	p.Plural.InitPluralClient()
	force := c.Bool("force")
	if err := crypto.CheckGitCrypt(c); err != nil {
		return errors.ErrorWrap(common.ErrNoGit, "Failed to scan your repo for secrets to encrypt them")
	}

	if c.IsSet("only") {
		installation, err := p.Plural.GetInstallation(c.String("only"))
		if err != nil {
			return api.GetErrorResponse(err, "GetInstallation")
		} else if installation == nil {
			return utils.HighlightError(fmt.Errorf("%s is not installed. Please install it with `plural bundle install`", c.String("only")))
		}

		return common.DoBuild(p.Plural.Client, installation, force)
	}

	installations, err := client.GetSortedInstallations(p.Plural, "")
	if err != nil {
		return err
	}

	for _, installation := range installations {
		if err := common.DoBuild(p.Plural.Client, installation, force); err != nil {
			return err
		}
	}
	return nil
}
