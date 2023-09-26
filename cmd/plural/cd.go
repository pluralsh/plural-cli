package plural

import (
	"github.com/pluralsh/plural/pkg/console"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/urfave/cli"
)

func init() {
	consoleToken = ""
	consoleURL = ""
}

var consoleToken string
var consoleURL string

func (p *Plural) cdCommands() []cli.Command {
	return []cli.Command{
		{
			Name:        "clusters",
			Subcommands: p.cdClusterCommands(),
			Usage:       "manage CD clusters",
		},
		{
			Name:        "services",
			Subcommands: p.cdServiceCommands(),
			Usage:       "manage CD services",
		},
		{
			Name:        "repositories",
			Subcommands: p.cdRepositoriesCommands(),
			Usage:       "manage CD repositories",
		},
	}
}

func (p *Plural) cdRepositoriesCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "list",
			Action: latestVersion(p.handleListCDRepositories),
			Usage:  "list repositories",
		},
		{
			Name:   "create",
			Action: latestVersion(p.handleCreateCDRepository),
			Usage:  "create repository",
		},
	}
}

func (p *Plural) cdServiceCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "list",
			ArgsUsage: "CLUSTER_ID",
			Action:    latestVersion(requireArgs(p.handleListClusterServices, []string{"CLUSTER_ID"})),
			Usage:     "list cluster services",
		},
	}
}

func (p *Plural) cdClusterCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "list",
			Action: latestVersion(p.handleListClusters),
			Usage:  "list clusters",
		},
	}
}

func (p *Plural) handleCreateCDRepository(c *cli.Context) error {

	return nil
}

func (p *Plural) handleListCDRepositories(c *cli.Context) error {

	return nil
}

func (p *Plural) handleListClusterServices(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	sd, err := p.ConsoleClient.ListClusterServices()
	if err != nil {
		return err
	}

	headers := []string{"Id", "Name", "Namespace", "Git URL", "Git Folder"}
	return utils.PrintTable(sd, headers, func(sd console.ServiceDeployment) ([]string, error) {
		return []string{sd.Id, sd.Name, sd.Namespace, sd.Git.Ref, sd.Git.Folder}, nil
	})

	return nil
}

func (p *Plural) handleListClusters(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	clusters, err := p.ConsoleClient.ListClusters()
	if err != nil {
		return err
	}

	headers := []string{"Id", "Name", "Version"}
	return utils.PrintTable(clusters, headers, func(cl console.Cluster) ([]string, error) {
		return []string{cl.Id, cl.Name, cl.Version}, nil
	})

	return nil
}
