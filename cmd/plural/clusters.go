package plural

import (
	"fmt"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/urfave/cli"
)

func (p *Plural) clusterCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "list",
			Usage:  "lists clusters accessible to your user",
			Action: latestVersion(p.listClusters),
		},
		{
			Name:  "view",
			Usage: "shows info for a cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "id",
					Usage: "the id of the source cluster",
				},
			},
			Action: latestVersion(p.showCluster),
		},
		{
			Name:  "depend",
			Usage: "have a cluster wait for promotion on another cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "source-id",
					Usage: "the id of the source cluster",
				},
				cli.StringFlag{
					Name:  "dest-id",
					Usage: "the id of the cluster waiting for promotion",
				},
			},
			Action: latestVersion(p.dependCluster),
		},
		{
			Name:   "promote",
			Usage:  "promote pending upgrades to your cluster",
			Action: latestVersion(p.promoteCluster),
		},
	}
}

func (p *Plural) listClusters(c *cli.Context) error {
	p.InitPluralClient()
	clusters, err := p.Client.Clusters()
	if err != nil {
		return err
	}

	headers := []string{"ID", "Name", "Provider", "Git Url", "Owner"}
	return utils.PrintTable(clusters, headers, func(c *api.Cluster) ([]string, error) {
		return []string{c.Id, c.Name, c.Provider, c.GitUrl, c.Owner.Email}, nil
	})
}

func (p *Plural) showCluster(c *cli.Context) error {
	p.InitPluralClient()
	id := c.String("id")
	if id == "" {
		clusters, err := p.Client.Clusters()
		if err != nil {
			return err
		}

		project, err := manifest.FetchProject()
		if err != nil {
			return err
		}
		for _, cluster := range clusters {
			if cluster.Name == project.Cluster && cluster.Owner.Email == project.Owner.Email {
				id = cluster.Id
				break
			}
		}
	}
	cluster, err := p.Client.Cluster(id)
	if err != nil {
		return err
	}

	fmt.Printf("Cluster %s:\n\n", cluster.Id)

	utils.PrintAttributes(map[string]string{
		"Id":       cluster.Id,
		"Name":     cluster.Name,
		"Provider": cluster.Provider,
		"Git Url":  cluster.GitUrl,
		"Owner":    cluster.Owner.Email,
	})

	fmt.Println("")
	if len(cluster.UpgradeInfo) > 0 {
		fmt.Printf("Pending Upgrades:\n\n")
		headers := []string{"Repository", "Count"}
		return utils.PrintTable(cluster.UpgradeInfo, headers, func(c *api.UpgradeInfo) ([]string, error) {
			return []string{c.Installation.Repository.Name, fmt.Sprintf("%d", c.Count)}, nil
		})
	}

	fmt.Println("No pending upgrades")
	return nil
}

func (p *Plural) dependCluster(c *cli.Context) error {
	p.InitPluralClient()
	source, dest := c.String("source-id"), c.String("dest-id")
	if err := p.Client.CreateDependency(source, dest); err != nil {
		return err
	}

	utils.Highlight("Cluster %s will now delegate upgrades to %s", dest, source)
	return nil
}

func (p *Plural) promoteCluster(c *cli.Context) error {
	p.InitPluralClient()
	if err := p.Client.PromoteCluster(); err != nil {
		return err
	}

	utils.Success("Upgrades promoted!")
	return nil
}
