package repo

import (
	"fmt"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/client"

	"github.com/pluralsh/plural-cli/pkg/common"

	"github.com/urfave/cli"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/bundle"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/format"
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
		Name:        "repos",
		Usage:       "view and manage plural repositories",
		Subcommands: p.reposCommands(),
		Category:    "API",
	}
}

func APICommand(clients client.Plural) cli.Command {
	p := Plural{
		Plural: clients,
	}
	return cli.Command{
		Name:        "apps",
		Usage:       "view and manage plural repositories",
		Subcommands: p.reposCommands(),
		Category:    "API",
	}
}

func (p *Plural) reposCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "unlock",
			Usage:     "unlocks installations in a repo that have breaking changes",
			ArgsUsage: "{app}",
			Action:    common.LatestVersion(common.RequireArgs(p.handleUnlockRepo, []string{"{app}"})),
		},
		{
			Name:      "release",
			Usage:     "tags the installations in the current cluster with the given release channels",
			ArgsUsage: "{app}",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "tag",
					Usage: "tag name for a given release channel, eg stable, warm, dev, prod",
				},
			},
			Action: common.LatestVersion(common.RequireArgs(p.handleRelease, []string{"{app}"})),
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
			Action: common.LatestVersion(p.handleResetInstallations),
		},
		{
			Name:   "synced",
			Usage:  "marks installations in this repo as being synced",
			Action: p.handleMarkSynced,
		},
		{
			Name:      "uninstall",
			Usage:     "uninstall an app from the plural api",
			ArgsUsage: "{app}",
			Action:    common.LatestVersion(common.RequireArgs(p.handleUninstall, []string{"{app}"})),
		},
		{
			Name:  "list",
			Usage: "list available repositories to install",
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
			Action: common.LatestVersion(p.handleListRepositories),
		},
	}
}

func (p *Plural) handleRelease(c *cli.Context) error {
	p.InitPluralClient()
	app := c.Args().First()
	tags := c.StringSlice("tag")
	err := p.Release(c.Args().First(), c.StringSlice("tag"))
	if err != nil {
		return api.GetErrorResponse(err, "Release")
	}

	utils.Success("Published release for %s to channels [%s]\n", app, strings.Join(tags, ", "))
	return nil
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

	if inst == nil {
		return fmt.Errorf("%s already uninstalled", c.Args().First())
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

func (p *Plural) handleMarkSynced(c *cli.Context) error {
	p.InitPluralClient()
	return p.MarkSynced(c.Args().Get(0))
}

func (p *Plural) handleResetInstallations(c *cli.Context) error {
	p.InitPluralClient()
	conf := config.Read()
	if !common.Confirm(fmt.Sprintf("Are you sure you want to reset installations for %s?  This will also wipe all oidc providers and any other associated state in the plural api", conf.Email), "PLURAL_REPOS_RESET_CONFIRM") {
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
