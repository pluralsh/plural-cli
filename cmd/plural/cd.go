package plural

import (
	"fmt"
	"strings"

	gqlclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/samber/lo"
	"github.com/urfave/cli"
	"k8s.io/apimachinery/pkg/util/yaml"
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
			Flags: []cli.Flag{
				cli.StringFlag{Name: "url", Usage: "git repo url"},
				cli.StringFlag{Name: "privateKey", Usage: "git repo private key"},
				cli.StringFlag{Name: "passphrase", Usage: "git repo passphrase"},
				cli.StringFlag{Name: "username", Usage: "git repo username"},
				cli.StringFlag{Name: "password", Usage: "git repo password"},
			},
			Usage: "create repository",
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
		{
			Name:      "create",
			ArgsUsage: "CLUSTER_ID",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "name", Usage: "service name"},
				cli.StringFlag{Name: "namespace", Usage: "service namespace"},
				cli.StringFlag{Name: "version", Usage: "service version"},
				cli.StringFlag{Name: "repo-id", Usage: "repository ID"},
				cli.StringFlag{Name: "git-ref", Usage: "git ref, can be branch, tag or commit sha"},
				cli.StringFlag{Name: "git-folder", Usage: "folder within the source tree where manifests are located"},
				cli.StringSliceFlag{
					Name:  "conf",
					Usage: "config name value",
				},
				cli.StringFlag{Name: "config-file", Usage: "path for configuration file"},
			},
			Action: latestVersion(requireArgs(p.handleCreateClusterService, []string{"CLUSTER_ID"})),
			Usage:  "create cluster service",
		},
		{
			Name:      "update",
			ArgsUsage: "SERVICE_ID",
			Action:    latestVersion(requireArgs(p.handleUpdateClusterService, []string{"SERVICE_ID"})),
			Usage:     "update cluster service",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "version", Usage: "service version"},
				cli.StringFlag{Name: "git-ref", Usage: "git ref, can be branch, tag or commit sha"},
				cli.StringFlag{Name: "git-folder", Usage: "folder within the source tree where manifests are located"},
				cli.StringSliceFlag{
					Name:  "conf",
					Usage: "config name value",
				},
			},
		},
		{
			Name:      "describe",
			ArgsUsage: "SERVICE_ID",
			Action:    latestVersion(requireArgs(p.handleDescribeClusterService, []string{"SERVICE_ID"})),
			Flags: []cli.Flag{
				cli.StringFlag{Name: "o", Usage: "output format"},
			},
			Usage: "describe cluster service",
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
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	repo, err := p.ConsoleClient.CreateRepository(c.String("url"), getFlag(c.String("privateKey")),
		getFlag(c.String("passphrase")), getFlag(c.String("username")), getFlag(c.String("password")))
	if err != nil {
		return err
	}

	headers := []string{"ID", "URL"}
	return utils.PrintTable([]gqlclient.GitRepositoryFragment{*repo.CreateGitRepository}, headers, func(r gqlclient.GitRepositoryFragment) ([]string, error) {
		return []string{r.ID, r.URL}, nil
	})
}

func (p *Plural) handleListCDRepositories(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	repos, err := p.ConsoleClient.ListRepositories()
	if err != nil {
		return err
	}

	headers := []string{"ID", "URL", "Status", "Error"}
	return utils.PrintTable(repos.GitRepositories.Edges, headers, func(r *gqlclient.GitRepositoryEdgeFragment) ([]string, error) {
		return []string{r.Node.ID, r.Node.URL, string(*r.Node.Health), lo.FromPtr(r.Node.Error)}, nil
	})

}

func (p *Plural) handleListClusterServices(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	clusterId := c.Args().Get(0)

	sd, err := p.ConsoleClient.ListClusterServices(clusterId)
	if err != nil {
		return err
	}

	headers := []string{"Id", "Name", "Namespace", "Git Ref", "Git Folder", "Repo ID"}
	return utils.PrintTable(sd.ServiceDeployments.Edges, headers, func(sd *gqlclient.ServiceDeploymentEdgeFragment) ([]string, error) {
		return []string{sd.Node.ID, sd.Node.Name, sd.Node.Namespace, sd.Node.Git.Ref, sd.Node.Git.Folder, sd.Node.Repository.ID}, nil
	})
}

type ServiceDeploymentAttributesConfiguration struct {
	Configuration []*gqlclient.ConfigAttributes
}

func (p *Plural) handleCreateClusterService(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	clusterId := c.Args().Get(0)
	v := c.String("version")
	attributes := gqlclient.ServiceDeploymentAttributes{
		Name:         c.String("name"),
		Namespace:    c.String("namespace"),
		Version:      &v,
		RepositoryID: c.String("repo-id"),
		Git: gqlclient.GitRefAttributes{
			Ref:    c.String("git-ref"),
			Folder: c.String("git-folder"),
		},
		Configuration: []*gqlclient.ConfigAttributes{},
	}

	if c.String("config-file") != "" {
		configFile, err := utils.ReadFile(c.String("config-file"))
		if err != nil {
			return err
		}
		sdc := ServiceDeploymentAttributesConfiguration{}
		if err := yaml.Unmarshal([]byte(configFile), &sdc); err != nil {
			return err
		}
		attributes.Configuration = append(attributes.Configuration, sdc.Configuration...)
	}
	var confArgs []string
	if c.IsSet("conf") {
		confArgs = append(confArgs, c.StringSlice("conf")...)
	}
	for _, conf := range confArgs {
		configurationPair := strings.Split(conf, "=")
		if len(configurationPair) == 2 {
			attributes.Configuration = append(attributes.Configuration, &gqlclient.ConfigAttributes{
				Name:  configurationPair[0],
				Value: configurationPair[1],
			})
		}
	}

	sd, err := p.ConsoleClient.CreateClusterService(clusterId, attributes)
	if err != nil {
		return err
	}
	if sd == nil {
		return fmt.Errorf("the returned object is empty, check if all fields are set")
	}

	headers := []string{"Id", "Name", "Namespace", "Git Ref", "Git Folder", "Repo"}
	return utils.PrintTable([]*gqlclient.CreateServiceDeployment{sd}, headers, func(sd *gqlclient.CreateServiceDeployment) ([]string, error) {
		return []string{sd.CreateServiceDeployment.ID, sd.CreateServiceDeployment.Name, sd.CreateServiceDeployment.Namespace, sd.CreateServiceDeployment.Git.Ref, sd.CreateServiceDeployment.Git.Folder, sd.CreateServiceDeployment.Repository.ID}, nil
	})
}

func (p *Plural) handleDescribeClusterService(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	serviceId := c.Args().Get(0)
	existing, err := p.ConsoleClient.GetClusterService(serviceId)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("existing service deployment is empty")
	}
	output := c.String("o")
	if output == "json" {
		utils.NewJsonPrinter(existing).PrettyPrint()
	} else {
		utils.NewYAMLPrinter(existing).PrettyPrint()
	}

	return nil
}

func (p *Plural) handleUpdateClusterService(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	serviceId := c.Args().Get(0)

	existing, err := p.ConsoleClient.GetClusterService(serviceId)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("existing service deployment is empty")
	}
	existingConfigurations := map[string]string{}
	attributes := gqlclient.ServiceUpdateAttributes{
		Version: &existing.ServiceDeployment.Version,
		Git: gqlclient.GitRefAttributes{
			Ref:    existing.ServiceDeployment.Git.Ref,
			Folder: existing.ServiceDeployment.Git.Folder,
		},
		Configuration: []*gqlclient.ConfigAttributes{},
	}

	for _, conf := range existing.ServiceDeployment.Configuration {
		existingConfigurations[conf.Name] = conf.Value
	}

	v := c.String("version")
	if v != "" {
		attributes.Version = &v
	}
	if c.String("git-ref") != "" {
		attributes.Git.Ref = c.String("git-ref")
	}
	if c.String("git-folder") != "" {
		attributes.Git.Folder = c.String("git-folder")
	}
	var confArgs []string
	if c.IsSet("conf") {
		confArgs = append(confArgs, c.StringSlice("conf")...)
	}

	updateConfigurations := map[string]string{}
	for _, conf := range confArgs {
		configurationPair := strings.Split(conf, "=")
		if len(configurationPair) == 2 {
			updateConfigurations[configurationPair[0]] = configurationPair[1]
		}
	}
	for k, v := range updateConfigurations {
		existingConfigurations[k] = v
	}
	for key, value := range existingConfigurations {
		attributes.Configuration = append(attributes.Configuration, &gqlclient.ConfigAttributes{
			Name:  key,
			Value: value,
		})
	}

	sd, err := p.ConsoleClient.UpdateClusterService(serviceId, attributes)
	if err != nil {
		return err
	}
	if sd == nil {
		return fmt.Errorf("returned object is nil")
	}

	headers := []string{"Id", "Name", "Namespace", "Git Ref", "Git Folder", "Repo"}
	return utils.PrintTable([]*gqlclient.UpdateServiceDeployment{sd}, headers, func(sd *gqlclient.UpdateServiceDeployment) ([]string, error) {
		return []string{sd.UpdateServiceDeployment.ID, sd.UpdateServiceDeployment.Name, sd.UpdateServiceDeployment.Namespace, sd.UpdateServiceDeployment.Git.Ref, sd.UpdateServiceDeployment.Git.Folder, sd.UpdateServiceDeployment.Repository.ID}, nil
	})
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
	return utils.PrintTable(clusters.Clusters.Edges, headers, func(cl *gqlclient.ClusterEdgeFragment) ([]string, error) {
		return []string{cl.Node.ID, cl.Node.Name, *cl.Node.Version}, nil
	})
}

func getFlag(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
