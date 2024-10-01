package bundle

import (
	"fmt"

	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/urfave/cli"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/bundle"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

type Plural struct {
	client.Plural
}

func Command(clients client.Plural) cli.Command {
	p := Plural{
		Plural: clients,
	}
	return cli.Command{
		Name:        "bundle",
		Usage:       "Commands for installing and discovering installation bundles",
		Subcommands: p.bundleCommands(),
	}
}

func (p *Plural) bundleCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "list",
			Usage:     "lists bundles for a repository",
			ArgsUsage: "{repo}",
			Action:    common.LatestVersion(common.Rooted(common.RequireArgs(p.bundleList, []string{"{repo}"}))),
		},
		{
			Name:      "install",
			Usage:     "installs a bundle and writes the configuration to this installation's context",
			ArgsUsage: "{repo} {bundle}",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "refresh",
					Usage: "re-enter the configuration for this bundle",
				},
			},
			Action: common.Tracked(common.LatestVersion(common.Rooted(p.bundleInstall)), "bundle.install"),
		},
	}
}

func (p *Plural) bundleList(c *cli.Context) error {
	repo := c.Args().Get(0)
	p.InitPluralClient()
	recipes, err := p.listRecipes(repo)
	if err != nil {
		return api.GetErrorResponse(err, "ListRecipes")
	}

	headers := []string{"Name", "Description", "Provider", "Install Command"}
	return utils.PrintTable(recipes, headers, func(recipe *api.Recipe) ([]string, error) {
		return []string{recipe.Name, recipe.Description, recipe.Provider, fmt.Sprintf("plural bundle install %s %s", repo, recipe.Name)}, nil
	})
}

func (p *Plural) bundleInstall(c *cli.Context) (err error) {
	args := c.Args()
	p.InitPluralClient()
	repo := args.Get(0)
	if repo == "" {
		return fmt.Errorf("REPO argument required, try running `plural bundle install REPO` for the app you want to install")
	}

	bdl := args.Get(1)
	if bdl == "" {
		recipes, err := p.listRecipes(args.Get(0))
		if err != nil {
			return err
		}
		for _, recipe := range recipes {
			if recipe.Primary {
				bdl = recipe.Name
				break
			}
		}

		if bdl == "" {
			return fmt.Errorf("you need to specify a bundle name, run `plural bundle list %s` to find eligible bundles then `plural bundle install %s <name>` to install", repo, repo)
		}
	}

	err = bundle.Install(p.Client, repo, bdl, c.Bool("refresh"))
	utils.Note("To edit the configuration you've just entered, edit the context.yaml file at the root of your repo, or run with the --refresh flag\n")
	return
}

func (p *Plural) listRecipes(repo string) (res []*api.Recipe, err error) {
	man, err := manifest.FetchProject()
	if err != nil {
		return
	}
	res, err = p.ListRecipes(repo, man.Provider)
	return
}
