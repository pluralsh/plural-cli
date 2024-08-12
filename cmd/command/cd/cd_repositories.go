package cd

import (
	"fmt"
	"github.com/pluralsh/plural-cli/pkg/common"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/samber/lo"
	"github.com/urfave/cli"
)

func (p *Plural) cdRepositories() cli.Command {
	return cli.Command{
		Name:        "repositories",
		Subcommands: p.cdRepositoriesCommands(),
		Usage:       "manage CD repositories",
	}
}

func (p *Plural) cdRepositoriesCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "list",
			Action: common.LatestVersion(p.handleListCDRepositories),
			Usage:  "list repositories",
		},
		{
			Name:   "create",
			Action: common.LatestVersion(p.handleCreateCDRepository),
			Flags: []cli.Flag{
				cli.StringFlag{Name: "url", Usage: "git repo url", Required: true},
				cli.StringFlag{Name: "private-key", Usage: "git repo private key"},
				cli.StringFlag{Name: "passphrase", Usage: "git repo passphrase"},
				cli.StringFlag{Name: "username", Usage: "git repo username"},
				cli.StringFlag{Name: "password", Usage: "git repo password"},
			},
			Usage: "create repository",
		},
		{
			Name:      "update",
			ArgsUsage: "REPO_ID",
			Action:    common.LatestVersion(common.RequireArgs(p.handleUpdateCDRepository, []string{"REPO_ID"})),
			Flags: []cli.Flag{
				cli.StringFlag{Name: "url", Usage: "git repo url", Required: true},
				cli.StringFlag{Name: "private-key", Usage: "git repo private key"},
				cli.StringFlag{Name: "passphrase", Usage: "git repo passphrase"},
				cli.StringFlag{Name: "username", Usage: "git repo username"},
				cli.StringFlag{Name: "password", Usage: "git repo password"},
			},
			Usage: "update repository",
		},
	}
}

func (p *Plural) handleListCDRepositories(_ *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	repos, err := p.ConsoleClient.ListRepositories()
	if err != nil {
		return err
	}
	if repos == nil {
		return fmt.Errorf("returned objects list [ListRepositories] is nil")
	}
	headers := []string{"ID", "URL", "Status", "Error"}
	return utils.PrintTable(repos.GitRepositories.Edges, headers, func(r *gqlclient.GitRepositoryEdgeFragment) ([]string, error) {
		health := "UNKNOWN"
		if r.Node.Health != nil {
			health = string(*r.Node.Health)
		}
		return []string{r.Node.ID, r.Node.URL, health, lo.FromPtr(r.Node.Error)}, nil
	})

}

func (p *Plural) handleCreateCDRepository(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	url := c.String("url")
	repo, err := p.ConsoleClient.CreateRepository(url, getFlag(c.String("privateKey")),
		getFlag(c.String("passphrase")), getFlag(c.String("username")), getFlag(c.String("password")))
	if err != nil {
		return err
	}

	headers := []string{"ID", "URL"}
	return utils.PrintTable([]gqlclient.GitRepositoryFragment{*repo.CreateGitRepository}, headers, func(r gqlclient.GitRepositoryFragment) ([]string, error) {
		return []string{r.ID, r.URL}, nil
	})
}

func (p *Plural) handleUpdateCDRepository(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	repoId := c.Args().Get(0)

	attr := gqlclient.GitAttributes{
		URL: c.String("url"),
	}

	if c.String("private-key") != "" {
		attr.PrivateKey = lo.ToPtr(c.String("private-key"))
	}

	if c.String("passphrase") != "" {
		attr.Passphrase = lo.ToPtr(c.String("passphrase"))
	}

	if c.String("password") != "" {
		attr.Password = lo.ToPtr(c.String("password"))
	}

	if c.String("username") != "" {
		attr.Username = lo.ToPtr(c.String("username"))
	}

	repo, err := p.ConsoleClient.UpdateRepository(repoId, attr)
	if err != nil {
		return err
	}

	headers := []string{"ID", "URL"}
	return utils.PrintTable([]gqlclient.GitRepositoryFragment{*repo.UpdateGitRepository}, headers, func(r gqlclient.GitRepositoryFragment) ([]string, error) {
		return []string{r.ID, r.URL}, nil
	})
}

func getFlag(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
