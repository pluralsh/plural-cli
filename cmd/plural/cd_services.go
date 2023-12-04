package plural

import (
	"fmt"
	"strings"

	gqlclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/samber/lo"
	"github.com/urfave/cli"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func (p *Plural) cdServices() cli.Command {
	return cli.Command{
		Name:        "services",
		Subcommands: p.cdServiceCommands(),
		Usage:       "manage CD services",
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
				cli.StringFlag{Name: "name", Usage: "service name", Required: true},
				cli.StringFlag{Name: "namespace", Usage: "service namespace. If not specified the 'default' will be used"},
				cli.StringFlag{Name: "version", Usage: "service version. If not specified the '0.0.1' will be used"},
				cli.StringFlag{Name: "repo-id", Usage: "repository ID", Required: true},
				cli.StringFlag{Name: "git-ref", Usage: "git ref, can be branch, tag or commit sha", Required: true},
				cli.StringFlag{Name: "git-folder", Usage: "folder within the source tree where manifests are located", Required: true},
				cli.StringFlag{Name: "kustomize-folder", Usage: "folder within the kustomize file is located"},
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
				cli.StringFlag{Name: "kustomize-folder", Usage: "folder within the kustomize file is located"},
				cli.StringSliceFlag{
					Name:  "conf",
					Usage: "config name value",
				},
			},
		},
		{
			Name:      "clone",
			ArgsUsage: "CLUSTER SERVICE",
			Action:    latestVersion(requireArgs(p.handleCloneClusterService, []string{"CLUSTER", "SERVICE"})),
			Flags: []cli.Flag{
				cli.StringFlag{Name: "name", Usage: "the name for the cloned service", Required: true},
				cli.StringFlag{Name: "namespace", Usage: "the namespace for this cloned service", Required: true},
				cli.StringSliceFlag{
					Name:  "conf",
					Usage: "config name value",
				},
			},
			Usage: "deep clone a service onto either the same cluster or another",
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
		{
			Name:      "delete",
			ArgsUsage: "SERVICE_ID",
			Action:    latestVersion(requireArgs(p.handleDeleteClusterService, []string{"SERVICE_ID"})),
			Usage:     "delete cluster service",
		},
	}
}

func (p *Plural) handleListClusterServices(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	sd, err := p.ConsoleClient.ListClusterServices(getIdAndName(c.Args().Get(0)))
	if err != nil {
		return err
	}
	if sd == nil {
		return fmt.Errorf("returned objects list [ListClusterServices] is nil")
	}
	headers := []string{"Id", "Name", "Namespace", "Git Ref", "Git Folder", "Repo"}
	return utils.PrintTable(sd, headers, func(sd *gqlclient.ServiceDeploymentEdgeFragment) ([]string, error) {
		return []string{sd.Node.ID, sd.Node.Name, sd.Node.Namespace, sd.Node.Git.Ref, sd.Node.Git.Folder, sd.Node.Repository.URL}, nil
	})
}

func (p *Plural) handleCreateClusterService(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	v, err := validateFlag(c, "version", "0.0.1")
	if err != nil {
		return err
	}
	name := c.String("name")
	namespace, err := validateFlag(c, "namespace", "default")
	if err != nil {
		return err
	}
	repoId := c.String("repo-id")
	gitRef := c.String("git-ref")
	gitFolder := c.String("git-folder")
	attributes := gqlclient.ServiceDeploymentAttributes{
		Name:         name,
		Namespace:    namespace,
		Version:      &v,
		RepositoryID: repoId,
		Git: &gqlclient.GitRefAttributes{
			Ref:    gitRef,
			Folder: gitFolder,
		},
		Configuration: []*gqlclient.ConfigAttributes{},
	}

	if c.String("kustomize-folder") != "" {
		attributes.Kustomize = &gqlclient.KustomizeAttributes{
			Path: c.String("kustomize-folder"),
		}
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
				Value: &configurationPair[1],
			})
		}
	}

	clusterId, clusterName := getIdAndName(c.Args().Get(0))
	sd, err := p.ConsoleClient.CreateClusterService(clusterId, clusterName, attributes)
	if err != nil {
		return err
	}
	if sd == nil {
		return fmt.Errorf("the returned object is empty, check if all fields are set")
	}

	headers := []string{"Id", "Name", "Namespace", "Git Ref", "Git Folder", "Repo"}
	return utils.PrintTable([]*gqlclient.ServiceDeploymentFragment{sd}, headers, func(sd *gqlclient.ServiceDeploymentFragment) ([]string, error) {
		return []string{sd.ID, sd.Name, sd.Namespace, sd.Git.Ref, sd.Git.Folder, sd.Repository.URL}, nil
	})
}

func (p *Plural) handleCloneClusterService(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	cluster, err := p.ConsoleClient.GetCluster(getIdAndName(c.Args().Get(0)))
	if err != nil {
		return err
	}
	if cluster == nil {
		return fmt.Errorf("could not find cluster %s", c.Args().Get(0))
	}

	serviceId, clusterName, serviceName, err := getServiceIdClusterNameServiceName(c.Args().Get(1))
	if err != nil {
		return err
	}

	attributes := gqlclient.ServiceCloneAttributes{
		Name:      c.String("name"),
		Namespace: lo.ToPtr(c.String("namespace")),
	}

	// TODO: DRY this up with service update
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
	for key, value := range updateConfigurations {
		attributes.Configuration = append(attributes.Configuration, &gqlclient.ConfigAttributes{
			Name:  key,
			Value: lo.ToPtr(value),
		})
	}

	sd, err := p.ConsoleClient.CloneService(cluster.ID, serviceId, serviceName, clusterName, attributes)
	if err != nil {
		return err
	}

	headers := []string{"Id", "Name", "Namespace", "Git Ref", "Git Folder", "Repo"}
	return utils.PrintTable([]*gqlclient.ServiceDeploymentFragment{sd}, headers, func(sd *gqlclient.ServiceDeploymentFragment) ([]string, error) {
		return []string{sd.ID, sd.Name, sd.Namespace, sd.Git.Ref, sd.Git.Folder, sd.Repository.URL}, nil
	})
}

func (p *Plural) handleUpdateClusterService(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	serviceId, clusterName, serviceName, err := getServiceIdClusterNameServiceName(c.Args().Get(0))
	if err != nil {
		return err
	}

	existing, err := p.ConsoleClient.GetClusterService(serviceId, serviceName, clusterName)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("existing service deployment is empty")
	}
	existingConfigurations := map[string]string{}
	attributes := gqlclient.ServiceUpdateAttributes{
		Version: &existing.Version,
		Git: &gqlclient.GitRefAttributes{
			Ref:    existing.Git.Ref,
			Folder: existing.Git.Folder,
		},
		Configuration: []*gqlclient.ConfigAttributes{},
	}
	if existing.Kustomize != nil {
		attributes.Kustomize = &gqlclient.KustomizeAttributes{
			Path: existing.Kustomize.Path,
		}
	}

	for _, conf := range existing.Configuration {
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
			Value: lo.ToPtr(value),
		})
	}
	if c.String("kustomize-folder") != "" {
		attributes.Kustomize = &gqlclient.KustomizeAttributes{
			Path: c.String("kustomize-folder"),
		}
	}
	sd, err := p.ConsoleClient.UpdateClusterService(serviceId, serviceName, clusterName, attributes)
	if err != nil {
		return err
	}
	if sd == nil {
		return fmt.Errorf("returned object is nil")
	}

	headers := []string{"Id", "Name", "Namespace", "Git Ref", "Git Folder", "Repo"}
	return utils.PrintTable([]*gqlclient.ServiceDeploymentFragment{sd}, headers, func(sd *gqlclient.ServiceDeploymentFragment) ([]string, error) {
		return []string{sd.ID, sd.Name, sd.Namespace, sd.Git.Ref, sd.Git.Folder, sd.Repository.URL}, nil
	})
}

func (p *Plural) handleDescribeClusterService(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	serviceId, clusterName, serviceName, err := getServiceIdClusterNameServiceName(c.Args().Get(0))
	if err != nil {
		return err
	}
	existing, err := p.ConsoleClient.GetClusterService(serviceId, serviceName, clusterName)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("existing service deployment is empty")
	}
	output := c.String("o")
	if output == "json" {
		utils.NewJsonPrinter(existing).PrettyPrint()
		return nil
	} else if output == "yaml" {
		utils.NewYAMLPrinter(existing).PrettyPrint()
		return nil
	}
	desc, err := console.DescribeService(existing)
	if err != nil {
		return err
	}
	fmt.Print(desc)

	return nil
}

func (p *Plural) handleDeleteClusterService(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	serviceId, clusterName, serviceName, err := getServiceIdClusterNameServiceName(c.Args().Get(0))
	if err != nil {
		return err
	}

	svc, err := p.ConsoleClient.GetClusterService(serviceId, serviceName, clusterName)
	if err != nil {
		return err
	}
	if svc == nil {
		return fmt.Errorf("Could not find service for %s", c.Args().Get(0))
	}

	deleted, err := p.ConsoleClient.DeleteClusterService(svc.ID)
	if err != nil {
		return fmt.Errorf("could not delete service: %w", err)
	}

	utils.Success("Service %s has been deleted successfully\n", deleted.DeleteServiceDeployment.Name)
	return nil
}

type ServiceDeploymentAttributesConfiguration struct {
	Configuration []*gqlclient.ConfigAttributes
}

func getServiceIdClusterNameServiceName(input string) (serviceId, clusterName, serviceName *string, err error) {
	if strings.HasPrefix(input, "@") {
		i := strings.Trim(input, "@")
		split := strings.Split(i, "/")
		if len(split) != 2 {
			err = fmt.Errorf("expected format @clusterName/serviceName")
			return
		}
		clusterName = &split[0]
		serviceName = &split[1]
	} else {
		serviceId = &input
	}
	return
}

func validateFlag(ctx *cli.Context, name string, defaultVal string) (string, error) {
	res := ctx.String(name)
	if res == "" {
		if defaultVal == "" {
			return "", fmt.Errorf("expected --%s flag", name)
		}
		res = defaultVal
	}

	return res, nil
}
