package provider

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/polly/algorithms"
	"github.com/pluralsh/polly/containers"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	providerapi "github.com/pluralsh/plural-cli/pkg/provider/api"
	"github.com/pluralsh/plural-cli/pkg/provider/gcp"
)

var cloudFlag bool
var clusterFlag string

type Providers struct {
	AvailableProviders []string
	Scaffolds          map[string]string
}

var (
	providers       = Providers{}
	filterProviders = containers.ToSet([]string{"GENERIC", "KIND", "LINODE", "EQUINIX"})
)

func GetProvider() (providerapi.Provider, error) {
	path := manifest.ProjectManifestPath()
	if project, err := manifest.ReadProject(path); err == nil {
		return FromManifest(project)
	}
	if err := getAvailableProviders(); err != nil {
		return nil, err
	}

	provider := ""
	prompt := &survey.Select{
		Message: "Select the cloud provider:",
		Options: providers.AvailableProviders,
	}
	if err := survey.AskOne(prompt, &provider, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}
	return New(provider)
}

func SetCloudFlag(cloud bool) {
	cloudFlag = cloud
}

func SetClusterFlag(cluster string) {
	clusterFlag = cluster
}

func FromManifest(man *manifest.ProjectManifest) (providerapi.Provider, error) {
	switch man.Provider {
	case api.ProviderGCP:
		return gcp.NewProvider(gcp.WithManifest(man))
	case api.ProviderAWS:
		return awsFromManifest(man)
	case api.ProviderAzure:
		return AzureFromManifest(man, nil)
	case api.TEST:
		return testFromManifest(man)
	default:
		return nil, fmt.Errorf("invalid provider name: %s", man.Provider)
	}
}

func New(provider string) (providerapi.Provider, error) {
	conf := config.Read()
	switch provider {
	case api.ProviderGCP:
		return gcp.NewProvider(gcp.WithConfig(conf, clusterFlag, cloudFlag))
	case api.ProviderAWS:
		return mkAWS(conf)
	case api.ProviderAzure:
		return mkAzure(conf)
	default:
		return nil, fmt.Errorf("invalid provider name: %s", provider)
	}
}

func getAvailableProviders() error {
	if providers.AvailableProviders == nil {
		client := api.NewClient()
		available, err := client.GetTfProviders()
		if err != nil {
			return api.GetErrorResponse(err, "GetTfProviders")
		}

		available = containers.ToSet(available).Difference(filterProviders).List()
		providers.AvailableProviders = algorithms.Map(available, strings.ToLower)
	}
	return nil
}
