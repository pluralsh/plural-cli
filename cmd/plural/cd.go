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
