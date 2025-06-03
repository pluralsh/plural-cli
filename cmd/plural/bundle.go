package plural

import (
	"fmt"
	"strings"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/bundle"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/urfave/cli"
)

func (p *Plural) bundleCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "list",
			Usage:     "lists bundles for a repository",
			ArgsUsage: "REPO",
			Action:    latestVersion(rooted(requireArgs(p.bundleList, []string{"repo"}))),
		},
		{
			Name:      "install",
			Usage:     "installs a bundle and writes the configuration to this installation's context",
			ArgsUsage: "REPO NAME",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "refresh",
					Usage: "re-enter the configuration for this bundle",
				},
			},
			Action: tracked(latestVersion(rooted(p.bundleInstall)), "bundle.install"),
		},
	}
}

func (p *Plural) stackCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "install",
			Usage:     "installs a plural stack for your current provider",
			ArgsUsage: "NAME",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "refresh",
					Usage: "re-enter the configuration for all bundles",
				},
			},
			Action: tracked(latestVersion(rooted(requireArgs(p.stackInstall, []string{"stack-name"}))), "stack.install"),
		},
		{
			Name:  "list",
			Usage: "lists stacks to potentially install",
			Flags: []cli.Flag{
				cli.BoolTFlag{
					Name:  "account",
					Usage: "only list stacks within your account",
				},
			},
			Action: latestVersion(rooted(p.stackList)),
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
	}

	err = bundle.Install(p.Client, repo, bdl, c.Bool("refresh"))
	utils.Note("To edit the configuration you've just entered, edit the context.yaml file at the root of your repo, or run with the --refresh flag\n")
	return
}

func (p *Plural) stackInstall(c *cli.Context) (err error) {
	name := c.Args().Get(0)
	man, err := manifest.FetchProject()
	if err != nil {
		return
	}

	p.InitPluralClient()
	err = bundle.Stack(p.Client, name, man.Provider, c.Bool("refresh"))
	utils.Note("To edit the configuration you've just entered, edit the context.yaml file at the root of your repo, or run with the --refresh flag\n")
	return
}

func (p *Plural) stackList(c *cli.Context) (err error) {
	p.InitPluralClient()
	stacks, err := p.ListStacks(c.Bool("account"))
	if err != nil {
		return api.GetErrorResponse(err, "ListStacks")
	}

	headers := []string{"Name", "Description", "Featured"}
	return utils.PrintTable(stacks, headers, func(s *api.Stack) ([]string, error) {
		return []string{s.Name, s.Description, fmt.Sprintf("%v", s.Featured)}, nil
	})
}

func (p *Plural) listRecipes(repo string) (res []*api.Recipe, err error) {
	man, err := manifest.FetchProject()
	prov := ""
	if err == nil {
		prov = strings.ToUpper(man.Provider)
	}

	res, err = p.ListRecipes(repo, prov)
	return
}
