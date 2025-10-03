package cd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/polly/algorithms"
	"github.com/pluralsh/polly/containers"
	"github.com/samber/lo"
	"github.com/urfave/cli"
	"sigs.k8s.io/yaml"

	"github.com/pluralsh/plural-cli/pkg/cd"
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/console/errors"
	"github.com/pluralsh/plural-cli/pkg/kubernetes/config"
	"github.com/pluralsh/plural-cli/pkg/utils"
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
			Action: common.LatestVersion(p.handleListClusters),
			Usage:  "list clusters",
		},
		{
			Name:      "describe",
			Action:    common.LatestVersion(common.RequireArgs(p.handleDescribeCluster, []string{"@{cluster-handle}"})),
			Usage:     "describe cluster",
			ArgsUsage: "@{cluster-handle}",
			Flags:     []cli.Flag{cli.StringFlag{Name: "o", Usage: "output format"}},
		},
		{
			Name:      "update",
			Action:    common.LatestVersion(common.RequireArgs(p.handleUpdateCluster, []string{"@{cluster-handle}"})),
			Usage:     "update cluster",
			ArgsUsage: "@{cluster-handle}",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "handle", Usage: "unique human readable name used to identify this cluster"},
				cli.StringFlag{Name: "kubeconf-path", Usage: "path to kubeconfig"},
				cli.StringFlag{Name: "kubeconf-context", Usage: "the kubeconfig context you want to use. If not specified, the current one will be used"},
			},
		},
		{
			Name:      "delete",
			Action:    common.LatestVersion(common.RequireArgs(p.handleDeleteCluster, []string{"@{cluster-handle}"})),
			Usage:     "deregisters a cluster in plural cd, and drains all services (unless --soft is specified)",
			ArgsUsage: "@{cluster-handle}",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "soft",
					Usage: "deletes a cluster in our system but doesn't drain resources, leaving them untouched",
				},
			},
		},
		{
			Name:      "get-credentials",
			Aliases:   []string{"kubeconfig"},
			Action:    common.LatestVersion(p.handleGetClusterCredentials),
			Usage:     "updates kubeconfig file with appropriate credentials to point to specified cluster",
			ArgsUsage: "@{cluster-handle}",
		},
		{
			Name:      "create",
			Action:    common.LatestVersion(common.RequireArgs(p.handleCreateCluster, []string{"{cluster-name}"})),
			Usage:     "create cluster",
			ArgsUsage: "{cluster-name}",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "handle", Usage: "unique human readable name used to identify this cluster"},
				cli.StringFlag{Name: "version", Usage: "kubernetes cluster version", Required: true},
			},
		},
		{
			Name:   "bootstrap",
			Action: common.LatestVersion(p.handleClusterBootstrap),
			Usage:  "creates a new BYOK cluster and installs the agent onto it using the current kubeconfig",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "name", Usage: "The name you'll give the cluster", Required: true},
				cli.StringFlag{Name: "handle", Usage: "optional handle for the cluster"},
				cli.StringFlag{Name: "values", Usage: "values file to use for the deployment agent helm chart", Required: false},
				cli.StringFlag{Name: "chart-loc", Usage: "URL or filepath of helm chart tar file. Use if not wanting to install helm chart from default plural repository.", Required: false},
				cli.StringFlag{Name: "project", Usage: "the project this cluster will belong to", Required: false},
				cli.StringSliceFlag{
					Name:  "tag",
					Usage: "a cluster tag to add, useful for targeting with global services",
				},
				cli.StringFlag{
					Name:  "metadata",
					Usage: "Path to metadata file, or '-' to read from stdin",
				},
			},
		},
		{
			Name:   "reinstall",
			Action: common.LatestVersion(p.handleClusterReinstall),
			Flags: []cli.Flag{
				cli.StringFlag{Name: "values", Usage: "values file to use for the deployment agent helm chart", Required: false},
				cli.StringFlag{Name: "chart-loc", Usage: "URL or filepath of helm chart tar file. Use if not wanting to install helm chart from default plural repository.", Required: false},
				cli.StringFlag{
					Name:  "metadata",
					Usage: "Path to metadata file, or '-' to read from stdin",
				},
			},
			Usage:     "reinstalls the deployment operator into a cluster",
			ArgsUsage: "@{cluster-handle}",
		},
	}
}

func (p *Plural) handleListClusters(_ *cli.Context) error {
	clusters, err := p.ListClusters()
	if err != nil {
		return err
	}
	headers := []string{"Id", "Name", "Handle", "Version", "Provider"}
	return utils.PrintTable(clusters, headers, func(cl *gqlclient.ClusterEdgeFragment) ([]string, error) {
		var distro gqlclient.ClusterDistro
		if cl.Node.Distro != nil {
			distro = *cl.Node.Distro
		}
		handle := ""
		if cl.Node.Handle != nil {
			handle = *cl.Node.Handle
		}
		version := ""
		if cl.Node.Version != nil {
			version = *cl.Node.CurrentVersion
		}
		return []string{cl.Node.ID, cl.Node.Name, handle, version, string(distro)}, nil
	})
}

func (p *Plural) ListClusters() ([]*gqlclient.ClusterEdgeFragment, error) {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return nil, err
	}
	clusters, err := p.ConsoleClient.ListClusters()
	if err != nil {
		return nil, err
	}
	if clusters == nil {
		return nil, fmt.Errorf("returned objects list [ListClusters] is nil")
	}
	return clusters.Clusters.Edges, nil
}

func (p *Plural) GetClusterId(handle string) (string, string, error) {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return "", "", err
	}

	existing, err := p.ConsoleClient.GetCluster(nil, lo.ToPtr(handle))
	if err != nil {
		return "", "", err
	}

	return existing.ID, existing.Name, nil
}

func (p *Plural) handleDescribeCluster(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	existing, err := p.ConsoleClient.GetCluster(common.GetIdAndName(c.Args().Get(0)))
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("existing cluster is empty")
	}
	output := c.String("o")
	switch output {
	case "json":
		utils.NewJsonPrinter(existing).PrettyPrint()
	case "yaml":
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
	existing, err := p.ConsoleClient.GetCluster(common.GetIdAndName(c.Args().Get(0)))
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
		if cl.Distro != nil {
			provider = string(*cl.Distro)
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

	existing, err := p.ConsoleClient.GetCluster(common.GetIdAndName(c.Args().Get(0)))
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("this cluster does not exist")
	}

	if c.Bool("soft") {
		fmt.Println("detaching cluster from Plural CD, this will leave all workloads running.")
		return p.ConsoleClient.DetachCluster(existing.ID)
	}

	return p.ConsoleClient.DeleteCluster(existing.ID)
}

func (p *Plural) handleGetClusterCredentials(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	handle := c.Args().Get(0)
	if handle == "" {
		clusters, err := p.ListClusters()
		if err != nil {
			return err
		}
		if len(clusters) == 0 {
			return fmt.Errorf("no clusters found")
		}

		prompt := &survey.Select{
			Message: "Select the cluster you want to get credentials for:",
			Options: algorithms.Map(clusters, func(cl *gqlclient.ClusterEdgeFragment) string {
				return cl.Node.Name
			}),
		}
		if err := survey.AskOne(prompt, &handle, survey.WithValidator(survey.Required)); err != nil {
			return err
		}
	}

	cluster, err := p.ConsoleClient.GetCluster(common.GetIdAndName(fmt.Sprintf("@%s", handle)))
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

func (p *Plural) handleClusterReinstall(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	handle := c.Args().Get(0)
	var err error
	if handle == "" {
		handle, err = utils.ReadLine("Enter the handle for the cluster you want to reinstall the agent in")
		if err != nil {
			return err
		}
	}

	id, name := common.GetIdAndName(handle)
	if c.IsSet("metadata") {
		jsonData, err := getMetadataJson(c.String("metadata"))
		if err != nil {
			return err
		}
		if jsonData == nil {
			return fmt.Errorf("metadata file is empty")
		}
		if cluster, err := p.ConsoleClient.GetCluster(id, name); err == nil {
			if _, err := p.ConsoleClient.UpdateCluster(cluster.ID, gqlclient.ClusterUpdateAttributes{
				Metadata: jsonData,
			}); err != nil {
				return err
			}
		}
	}

	return p.ReinstallOperator(c, id, name, c.String("chart-loc"))
}

func (p *Plural) handleClusterBootstrap(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	attrs := gqlclient.ClusterAttributes{Name: c.String("name")}
	if c.String("handle") != "" {
		attrs.Handle = lo.ToPtr(c.String("handle"))
	}

	if c.String("project") != "" {
		project, err := p.ConsoleClient.GetProject(c.String("project"))
		if err != nil {
			return err
		}
		if project == nil {
			return fmt.Errorf("could not find project %s", c.String("project"))
		}

		attrs.ProjectID = lo.ToPtr(project.ID)
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
	if c.IsSet("metadata") {
		jsonData, err := getMetadataJson(c.String("metadata"))
		if err != nil {
			return err
		}
		if jsonData == nil {
			return fmt.Errorf("metadata file is empty")
		}
		attrs.Metadata = jsonData
	}
	existing, err := p.ConsoleClient.CreateCluster(attrs)
	if err != nil {
		if errors.Like(err, "handle") && common.Affirm("Do you want to reinstall the deployment operator?", "PLURAL_INSTALL_AGENT_CONFIRM_IF_EXISTS") {
			handle := lo.ToPtr(attrs.Name)
			if attrs.Handle != nil {
				handle = attrs.Handle
			}
			return p.ReinstallOperator(c, nil, handle, c.String("chart-loc"))
		}

		return err
	}

	if existing.CreateCluster.DeployToken == nil {
		return fmt.Errorf("could not fetch deploy token from cluster")
	}

	url := p.ConsoleClient.ExtUrl()
	if agentUrl, err := p.ConsoleClient.AgentUrl(existing.CreateCluster.ID); err == nil {
		url = agentUrl
	}

	deployToken := *existing.CreateCluster.DeployToken
	utils.Highlight("installing agent on %s with url %s\n", c.String("name"), p.ConsoleClient.Url())
	return p.DoInstallOperator(url, deployToken, c.String("values"), c.String("chart-loc"))
}

func getMetadataJson(val string) (*string, error) {
	var reader io.Reader
	if val == "-" {
		reader = os.Stdin
	} else {
		f, err := os.Open(val)
		if err != nil {
			return nil, err
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				return
			}
		}(f)
		reader = f
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, err
	}
	return lo.ToPtr(string(jsonData)), nil
}
