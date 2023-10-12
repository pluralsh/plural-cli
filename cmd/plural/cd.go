package plural

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gqlclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/plural/pkg/console"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/samber/lo"
	"github.com/urfave/cli"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
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
		{
			Name:   "install",
			Action: p.handleInstallDeploymentsOperator,
			Usage:  "install deployments operator",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "url", Usage: "console url", Required: true},
				cli.StringFlag{Name: "token", Usage: "console token", Required: true},
			},
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
				cli.StringFlag{Name: "url", Usage: "git repo url", Required: true},
				cli.StringFlag{Name: "privateKey", Usage: "git repo private key"},
				cli.StringFlag{Name: "passphrase", Usage: "git repo passphrase"},
				cli.StringFlag{Name: "username", Usage: "git repo username"},
				cli.StringFlag{Name: "password", Usage: "git repo password"},
			},
			Usage: "create repository",
		},
		{
			Name:      "update",
			ArgsUsage: "REPO_ID",
			Action:    latestVersion(requireArgs(p.handleUpdateCDRepository, []string{"REPO_ID"})),
			Flags: []cli.Flag{
				cli.StringFlag{Name: "url", Usage: "git repo url", Required: true},
			},
			Usage: "update repository",
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
				cli.StringFlag{Name: "name", Usage: "service name", Required: true},
				cli.StringFlag{Name: "namespace", Usage: "service namespace. If not specified the 'default' will be used"},
				cli.StringFlag{Name: "version", Usage: "service version. If not specified the '0.0.1' will be used"},
				cli.StringFlag{Name: "repo-id", Usage: "repository ID", Required: true},
				cli.StringFlag{Name: "git-ref", Usage: "git ref, can be branch, tag or commit sha", Required: true},
				cli.StringFlag{Name: "git-folder", Usage: "folder within the source tree where manifests are located", Required: true},
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
		{
			Name:      "delete",
			ArgsUsage: "SERVICE_ID",
			Action:    latestVersion(requireArgs(p.handleDeleteClusterService, []string{"SERVICE_ID"})),
			Usage:     "delete cluster service",
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
		{
			Name:      "describe",
			Action:    latestVersion(requireArgs(p.handleDescribeCluster, []string{"CLUSTER_ID"})),
			Usage:     "describe cluster",
			ArgsUsage: "CLUSTER_ID",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "o", Usage: "output format"},
			},
		},
		{
			Name:      "update",
			Action:    latestVersion(requireArgs(p.handleUpdateCluster, []string{"CLUSTER_ID"})),
			Usage:     "update cluster",
			ArgsUsage: "CLUSTER_ID",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "handle", Usage: "unique human readable name used to identify this cluster"},
				cli.StringFlag{Name: "kubeconf-path", Usage: "path to kubeconfig"},
				cli.StringFlag{Name: "kubeconf-context", Usage: "the kubeconfig context you want to use. If not specified, the current one will be used"},
			},
		},
	}
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

	repo, err := p.ConsoleClient.UpdateRepository(repoId, attr)
	if err != nil {
		return err
	}

	headers := []string{"ID", "URL"}
	return utils.PrintTable([]gqlclient.GitRepositoryFragment{*repo.UpdateGitRepository}, headers, func(r gqlclient.GitRepositoryFragment) ([]string, error) {
		return []string{r.ID, r.URL}, nil
	})
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
		return []string{r.Node.ID, r.Node.URL, string(*r.Node.Health), lo.FromPtr(r.Node.Error)}, nil
	})

}

func getIdAndName(input string) (id, name *string) {
	if strings.HasPrefix(input, "@") {
		h := strings.Trim(input, "@")
		name = &h
	} else {
		id = &input
	}
	return
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

type ServiceDeploymentAttributesConfiguration struct {
	Configuration []*gqlclient.ConfigAttributes
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
		Git: gqlclient.GitRefAttributes{
			Ref:    gitRef,
			Folder: gitFolder,
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
			Value: &value,
		})
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

func (p *Plural) handleListClusters(_ *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	clusters, err := p.ConsoleClient.ListClusters()
	if err != nil {
		return err
	}
	if clusters == nil {
		return fmt.Errorf("returned objects list [ListClusters] is nil")
	}
	headers := []string{"Id", "Name", "Handle", "Version", "Provider"}
	return utils.PrintTable(clusters.Clusters.Edges, headers, func(cl *gqlclient.ClusterEdgeFragment) ([]string, error) {
		provider := ""
		if cl.Node.Provider != nil {
			provider = cl.Node.Provider.Name
		}
		handle := ""
		if cl.Node.Handle != nil {
			handle = *cl.Node.Handle
		}
		return []string{cl.Node.ID, cl.Node.Name, handle, *cl.Node.Version, provider}, nil
	})
}

func (p *Plural) handleInstallDeploymentsOperator(c *cli.Context) error {
	namespace := "plrl-deploy-operator"
	err := p.InitKube()
	if err != nil {
		return err
	}
	err = p.Kube.CreateNamespace(namespace)
	if !apierrors.IsAlreadyExists(err) {
		return err
	}
	return console.InstallAgent(c.String("url"), c.String("token"), namespace)
}

func (p *Plural) handleDeleteClusterService(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	serviceId := c.Args().Get(0)
	existing, err := p.ConsoleClient.DeleteClusterService(serviceId)
	if err != nil {
		return fmt.Errorf("could not delete service: %s", err)
	}

	utils.Success("Service %s/%s has been deleted successfully", existing.DeleteServiceDeployment.ID, existing.DeleteServiceDeployment.Name)
	return nil
}

func (p *Plural) handleDescribeCluster(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	existing, err := p.ConsoleClient.GetCluster(getIdAndName(c.Args().Get(0)))
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("existing cluster is empty")
	}
	output := c.String("o")
	if output == "json" {
		utils.NewJsonPrinter(existing).PrettyPrint()
	} else if output == "yaml" {
		utils.NewYAMLPrinter(existing).PrettyPrint()
	}
	desc, err := console.DescribeCluster(existing)
	if err != nil {
		return err
	}
	fmt.Print(desc)
	return nil
}

func (p *Plural) handleUpdateCluster(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	existing, err := p.ConsoleClient.GetCluster(getIdAndName(c.Args().Get(0)))
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("existing cluster is empty")
	}
	updateAttr := gqlclient.ClusterUpdateAttributes{
		Version: *existing.Version,
		Handle:  existing.Handle,
	}
	newHandle := c.String("handle")
	if newHandle != "" {
		updateAttr.Handle = &newHandle
	}
	kubeconfigPath := c.String("kubeconf-path")
	if kubeconfigPath != "" {
		kubeconfig, err := getKubeconfig(kubeconfigPath, c.String("kubeconf-context"))
		if err != nil {
			return err
		}

		updateAttr.Kubeconfig = &gqlclient.KubeconfigAttributes{
			Raw: &kubeconfig,
		}
	}

	result, err := p.ConsoleClient.UpdateCluster(existing.ID, updateAttr)
	if err != nil {
		return err
	}
	headers := []string{"Id", "Name", "Handle", "Version", "Provider"}
	return utils.PrintTable([]gqlclient.ClusterFragment{*result.UpdateCluster}, headers, func(cl gqlclient.ClusterFragment) ([]string, error) {
		provider := ""
		if cl.Provider != nil {
			provider = cl.Provider.Name
		}
		handle := ""
		if cl.Handle != nil {
			handle = *cl.Handle
		}
		return []string{cl.ID, cl.Name, handle, *cl.Version, provider}, nil
	})

	return nil
}

func getKubeconfig(path, context string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, path[2:])
	}
	if !utils.Exists(path) {
		return "", fmt.Errorf("the specified path does not exist")
	}

	config, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return "", err
	}

	if context != "" {
		if config.Contexts[context] == nil {
			return "", fmt.Errorf("the given context doesn't exist")
		}
		config.CurrentContext = context
	}
	newConfig := *clientcmdapi.NewConfig()
	newConfig.CurrentContext = config.CurrentContext
	newConfig.Contexts[config.CurrentContext] = config.Contexts[config.CurrentContext]
	newConfig.Clusters[config.CurrentContext] = config.Clusters[config.CurrentContext]
	newConfig.AuthInfos[config.CurrentContext] = config.AuthInfos[config.CurrentContext]
	newConfig.Extensions[config.CurrentContext] = config.Extensions[config.CurrentContext]
	newConfig.Preferences = config.Preferences
	result, err := clientcmd.Write(newConfig)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

func getFlag(s string) *string {
	if s == "" {
		return nil
	}
	return &s
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
