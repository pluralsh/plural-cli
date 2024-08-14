package bounce

import (
	"os"
	"path/filepath"

	"github.com/pluralsh/plural-cli/pkg/utils/git"
	"github.com/pluralsh/plural-cli/pkg/utils/pathing"
	"github.com/pluralsh/plural-cli/pkg/wkspace"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/utils"
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
		Name:      "bounce",
		Aliases:   []string{"b"},
		Usage:     "redeploys the charts in a workspace",
		ArgsUsage: "APP",
		Action:    common.LatestVersion(common.InitKubeconfig(common.Owned(p.bounce))),
	}
}

func (p *Plural) bounce(c *cli.Context) error {
	p.Plural.InitPluralClient()
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}
	repoName := c.Args().Get(0)

	if repoName != "" {
		installation, err := p.Plural.GetInstallation(repoName)
		if err != nil {
			return api.GetErrorResponse(err, "GetInstallation")
		}
		return p.doBounce(repoRoot, installation)
	}

	installations, err := client.GetSortedInstallations(p.Plural, repoName)
	if err != nil {
		return err
	}

	for _, installation := range installations {
		if err := p.doBounce(repoRoot, installation); err != nil {
			return err
		}
	}
	return nil
}

func (p *Plural) doBounce(repoRoot string, installation *api.Installation) error {
	p.Plural.InitPluralClient()
	repoName := installation.Repository.Name
	utils.Warn("bouncing deployments in %s\n", repoName)
	workspace, err := wkspace.New(p.Plural.Client, installation)
	if err != nil {
		return err
	}

	if err := os.Chdir(pathing.SanitizeFilepath(filepath.Join(repoRoot, repoName))); err != nil {
		return err
	}
	return workspace.Bounce()
}
