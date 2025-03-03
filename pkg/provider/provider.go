package provider

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/provider/permissions"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/polly/algorithms"
	"github.com/pluralsh/polly/containers"
)

var cloudFlag bool
var clusterFlag string

type Provider interface {
	Name() string
	Cluster() string
	Project() string
	Region() string
	Bucket() string
	KubeConfig() error
	KubeContext() string
	CreateBucket() error
	Context() map[string]interface{}
	Preflights() []*Preflight
	Permissions() (permissions.Checker, error)
	Flush() error
}

type Preflight struct {
	Name     string
	Callback func() error
}

func (pf *Preflight) Validate() error {
	utils.Highlight("Executing preflight check :: %s ", pf.Name)
	if err := pf.Callback(); err != nil {
		fmt.Println("\nFound error:")
		return err
	}

	utils.Success("\u2713\n")
	return nil
}

type Providers struct {
	AvailableProviders []string
	Scaffolds          map[string]string
}

var (
	providers       = Providers{}
	filterProviders = containers.ToSet([]string{"GENERIC", "KIND", "LINODE", "EQUINIX"})
)

func GetProvider() (Provider, error) {
	path := manifest.ProjectManifestPath()
	if project, err := manifest.ReadProject(path); err == nil {
		return FromManifest(project)
	}
	if err := getAvailableProviders(); err != nil {
		return nil, err
	}

	provider := ""
	prompt := &survey.Select{
		Message: "Select one of the following providers:",
		Options: providers.AvailableProviders,
	}
	if err := survey.AskOne(prompt, &provider, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}
	utils.Success("Using provider %s\n", provider)
	return New(provider)
}

func SetCloudFlag(cloud bool) {
	cloudFlag = cloud
}

func SetClusterFlag(cluster string) {
	clusterFlag = cluster
}

func FromManifest(man *manifest.ProjectManifest) (Provider, error) {
	switch man.Provider {
	case api.ProviderGCP:
		return gcpFromManifest(man)
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

func New(provider string) (Provider, error) {
	conf := config.Read()
	switch provider {
	case api.ProviderGCP:
		return mkGCP(conf)
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
