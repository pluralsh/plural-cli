package plural

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	gqlclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/plural-cli/pkg/cd"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/kubernetes/config"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/polly/containers"
	"github.com/samber/lo"
	"github.com/urfave/cli"
)

var providerSurvey = []*survey.Question{
	{
		Name:   "name",
		Prompt: &survey.Input{Message: "Enter the name of your provider:"},
	},
	{
		Name:   "namespace",
		Prompt: &survey.Input{Message: "Enter the namespace of your provider:"},
	},
}

func (p *Plural) cdClusters() cli.Command {
	return cli.Command{
		Name:        "clusters",
		Subcommands: p.cdClusterCommands(),
		Usage:       "manage CD clusters",
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
		{
			Name:      "delete",
			Action:    latestVersion(requireArgs(p.handleDeleteCluster, []string{"CLUSTER_ID"})),
			Usage:     "deregisters a cluster in plural cd, and drains all services (unless --soft is specified)",
			ArgsUsage: "CLUSTER_ID",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "soft", Usage: "deletes a cluster in our system but doesn't drain resources, leaving them untouched"},
			},
		},
		{
			Name:      "get-credentials",
			Aliases:   []string{"kubeconfig"},
			Action:    latestVersion(requireArgs(p.handleGetClusterCredentials, []string{"CLUSTER_ID"})),
			Usage:     "updates kubeconfig file with appropriate credentials to point to specified cluster",
			ArgsUsage: "CLUSTER_ID",
		},
		{
			Name:      "create",
			Action:    latestVersion(requireArgs(p.handleCreateCluster, []string{"CLUSTER_NAME"})),
			Usage:     "create cluster",
			ArgsUsage: "CLUSTER_NAME",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "handle", Usage: "unique human readable name used to identify this cluster"},
				cli.StringFlag{Name: "version", Usage: "kubernetes cluster version", Required: true},
			},
		},
		{
			Name:   "bootstrap",
			Action: latestVersion(p.handleClusterBootstrap),
			Usage:  "creates a new BYOK cluster and installs the agent onto it using the current kubeconfig",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "name", Usage: "The name you'll give the cluster", Required: true},
				cli.StringFlag{Name: "handle", Usage: "optional handle for the cluster"},
				cli.StringSliceFlag{
					Name:  "tag",
					Usage: "a cluster tag to add, useful for targeting with global services",
				},
			},
		},
	}
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
		version := ""
		if cl.Node.Version != nil {
			version = *cl.Node.Version
		}
		return []string{cl.Node.ID, cl.Node.Name, handle, version, provider}, nil
	})
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
		return fmt.Errorf("this cluster does not exist")
	}
	updateAttr := gqlclient.ClusterUpdateAttributes{
		Version: existing.Version,
		Handle:  existing.Handle,
	}
	newHandle := c.String("handle")
	if newHandle != "" {
		updateAttr.Handle = &newHandle
	}
	kubeconfigPath := c.String("kubeconf-path")
	if kubeconfigPath != "" {
		kubeconfig, err := config.GetKubeconfig(kubeconfigPath, c.String("kubeconf-context"))
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
}

func (p *Plural) handleDeleteCluster(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	existing, err := p.ConsoleClient.GetCluster(getIdAndName(c.Args().Get(0)))
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("this cluster does not exist")
	}

	return p.ConsoleClient.DeleteCluster(existing.ID)
}
func (p *Plural) handleGetClusterCredentials(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	cluster, err := p.ConsoleClient.GetCluster(getIdAndName(c.Args().Get(0)))
	if err != nil {
		return err
	}
	if cluster == nil {
		return fmt.Errorf("cluster is nil")
	}

	return cd.SaveClusterKubeconfig(cluster, p.ConsoleClient.Token())
}

func (p *Plural) handleCreateCluster(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	name := c.Args().Get(0)
	attr := gqlclient.ClusterAttributes{
		Name: name,
	}
	if c.String("handle") != "" {
		attr.Handle = lo.ToPtr(c.String("handle"))
	}
	if c.String("version") != "" {
		attr.Version = lo.ToPtr(c.String("version"))
	}

	providerList, err := p.ConsoleClient.ListProviders()
	if err != nil {
		return err
	}
	providerNames := []string{}
	providerMap := map[string]string{}
	cloudProviders := []string{}
	for _, prov := range providerList.ClusterProviders.Edges {
		providerNames = append(providerNames, prov.Node.Name)
		providerMap[prov.Node.Name] = prov.Node.ID
		cloudProviders = append(cloudProviders, prov.Node.Cloud)
	}

	existingProv := containers.ToSet[string](cloudProviders)
	availableProv := containers.ToSet[string](availableProviders)
	toCreate := availableProv.Difference(existingProv)
	createNewProvider := "Create New Provider"

	if toCreate.Len() != 0 {
		providerNames = append(providerNames, createNewProvider)
	}

	prompt := &survey.Select{
		Message: "Select one of the following providers:",
		Options: providerNames,
	}
	provider := ""
	if err := survey.AskOne(prompt, &provider, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	if provider != createNewProvider {
		utils.Success("Using provider %s\n", provider)
		id := providerMap[provider]
		attr.ProviderID = &id
	} else {

		clusterProv, err := p.handleCreateProvider(toCreate.List())
		if err != nil {
			return err
		}
		if clusterProv == nil {
			utils.Success("All supported providers are created\n")
			return nil
		}
		utils.Success("Provider %s created successfully\n", clusterProv.CreateClusterProvider.Name)
		attr.ProviderID = &clusterProv.CreateClusterProvider.ID
		provider = clusterProv.CreateClusterProvider.Cloud
	}

	ca, err := cd.AskCloudSettings(provider)
	if err != nil {
		return err
	}
	attr.CloudSettings = ca

	existing, err := p.ConsoleClient.CreateCluster(attr)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("couldn't create cluster")
	}
	return nil
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

func (p *Plural) handleCreateProvider(existingProviders []string) (*gqlclient.CreateClusterProvider, error) {
	provider := ""
	var resp struct {
		Name      string
		Namespace string
	}
	if err := survey.Ask(providerSurvey, &resp); err != nil {
		return nil, err
	}

	prompt := &survey.Select{
		Message: "Select one of the following providers:",
		Options: existingProviders,
	}
	if err := survey.AskOne(prompt, &provider, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}

	cps, err := cd.AskCloudProviderSettings(provider)
	if err != nil {
		return nil, err
	}

	providerAttr := gqlclient.ClusterProviderAttributes{
		Name:          resp.Name,
		Namespace:     &resp.Namespace,
		Cloud:         &provider,
		CloudSettings: cps,
	}
	clusterProv, err := p.ConsoleClient.CreateProvider(providerAttr)
	if err != nil {
		return nil, err
	}
	if clusterProv == nil {
		return nil, fmt.Errorf("provider was not created properly")
	}
	return clusterProv, nil
}

func (p *Plural) handleClusterBootstrap(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	attrs := gqlclient.ClusterAttributes{Name: c.String("name")}
	if c.String("handle") != "" {
		attrs.Handle = lo.ToPtr(c.String("handle"))
	}

	if c.IsSet("tag") {
		attrs.Tags = lo.Map(c.StringSlice("tag"), func(tag string, index int) *gqlclient.TagAttributes {
			tags := strings.Split(tag, "=")
			if len(tags) == 2 {
				return &gqlclient.TagAttributes{
					Name:  tags[0],
					Value: tags[1],
				}
			}
			return nil
		})
		attrs.Tags = lo.Filter(attrs.Tags, func(t *gqlclient.TagAttributes, ind int) bool { return t != nil })
	}

	existing, err := p.ConsoleClient.CreateCluster(attrs)
	if err != nil {
		return err
	}

	if existing.CreateCluster.DeployToken == nil {
		return fmt.Errorf("could not fetch deploy token from cluster")
	}

	deployToken := *existing.CreateCluster.DeployToken
	url := fmt.Sprintf("%s/ext/gql", p.ConsoleClient.Url())
	utils.Highlight("instaling agent on %s with url %s and initial deploy token %s\n", c.String("name"), p.ConsoleClient.Url(), deployToken)
	return p.doInstallOperator(url, deployToken)
}
