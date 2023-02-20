package main

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/bundle"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/format"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils"
)

func (p *Plural) reposCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "unlock",
			Usage:     "unlocks installations in a repo that have breaking changes",
			ArgsUsage: "REPO",
			Action:    latestVersion(p.handleUnlockRepo),
		},
		{
			Name:  "reinstall",
			Usage: "reinstalls all bundles from a previous installation",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "refresh",
					Usage: "re-enter the configuration for all bundles",
				},
			},
			Action: p.handleReinstall,
		},
		{
			Name:   "reset",
			Usage:  "eliminates your current plural installation set, to change cloud provider or eject from plural",
			Action: latestVersion(p.handleResetInstallations),
		},
		{
			Name:      "uninstall",
			Usage:     "uninstall an app from the plural api",
			ArgsUsage: "APP",
			Action:    latestVersion(requireArgs(p.handleUninstall, []string{"APP"})),
		},
		{
			Name:      "list",
			Usage:     "list available repositories to install",
			ArgsUsage: "",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "query",
					Usage: "string to search by",
				},
				cli.StringFlag{
					Name:  "format",
					Usage: "format to print the repositories out, eg csv or default is table",
				},
			},
			Action: latestVersion(p.handleListRepositories),
		},
	}
}

func (p *Plural) handleUnlockRepo(c *cli.Context) error {
	p.InitPluralClient()
	err := p.UnlockRepository(c.Args().First())
	return api.GetErrorResponse(err, "UnlockRepository")
}

func (p *Plural) handleUninstall(c *cli.Context) error {
	p.InitPluralClient()
	inst, err := p.GetInstallation(c.Args().First())
	if err != nil {
		return api.GetErrorResponse(err, "GetInstallation")
	}

	err = p.DeleteInstallation(inst.Id)
	return api.GetErrorResponse(err, "DeleteInstallation")
}

func (p *Plural) handleListRepositories(c *cli.Context) error {
	p.InitPluralClient()
	repos, err := p.ListRepositories(c.String("query"))
	if err != nil {
		return api.GetErrorResponse(err, "ListRepositories")
	}

	addIcon := c.String("format") == "csv"

	formatter := format.New(format.FormatType(c.String("format")))
	header := []string{"Repo", "Description", "Publisher", "Bundles"}
	if addIcon {
		header = append(header, "Icon")
	}

	formatter.Header(header)
	for _, repo := range repos {
		recipeNames := utils.Map(repo.Recipes, func(recipe *api.Recipe) string {
			return recipe.Name
		})

		line := []string{repo.Name, repo.Description, repo.Publisher.Name, strings.Join(recipeNames, ", ")}
		if addIcon {
			line = append(line, repo.Icon)
		}
		if err := formatter.Write(line); err != nil {
			return err
		}
	}

	if err := formatter.Flush(); err != nil {
		return err
	}
	return nil
}

func (p *Plural) handleReinstall(c *cli.Context) error {
	p.InitPluralClient()
	ctx, err := manifest.FetchContext()
	if err != nil {
		return err
	}

	for _, b := range ctx.Bundles {
		if err := bundle.Install(p.Client, b.Repository, b.Name, c.Bool("refresh")); err != nil {
			return err
		}

		fmt.Println("Moving to the next bundle....")
	}

	return nil
}

func (p *Plural) handleResetInstallations(c *cli.Context) error {
	p.InitPluralClient()
	conf := config.Read()
	if !confirm(fmt.Sprintf("Are you sure you want to reset installations for %s?  This will also wipe all oidc providers and any other associated state in the plural api", conf.Email), "PLURAL_REPOS_RESET_CONFIRM") {
		return nil
	}

	count, err := p.ResetInstallations()
	if err != nil {
		return api.GetErrorResponse(err, "ResetInstallations")
	}

	fmt.Printf("Deleted %d installations in app.plural.sh\n", count)
	fmt.Println("(you can recreate these at any time and any running infrastructure is not affected, plural will simply no longer deliver upgrades)")
	utils.Note("Now run `plural bundle install <repo> <bundle-name>` to install a new app \n")
	return nil
}
