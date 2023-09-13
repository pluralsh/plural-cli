package provider

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/polly/algorithms"
	"github.com/pluralsh/polly/containers"
	v1 "k8s.io/api/core/v1"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/provider/permissions"
	"github.com/pluralsh/plural/pkg/utils"
)

type Provider interface {
	Name() string
	Cluster() string
	Project() string
	Region() string
	Bucket() string
	KubeConfig() error
	KubeContext() string
	CreateBackend(prefix string, version string, ctx map[string]interface{}) (string, error)
	Context() map[string]interface{}
	Decommision(node *v1.Node) error
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
	filterProviders = containers.ToSet([]string{"GENERIC", "KIND"})
)

func GetProviderScaffold(provider, version string) (string, error) {
	if providers.Scaffolds == nil {
		providers.Scaffolds = make(map[string]string)
	}
	_, ok := providers.Scaffolds[provider]
	if !ok {
		client := api.NewClient()
		scaffold, err := client.GetTfProviderScaffold(provider, version)
		providers.Scaffolds[provider] = scaffold
		if err != nil {
			return "", api.GetErrorResponse(err, "GetTfProviderScaffold")
		}
	}
	return providers.Scaffolds[provider], nil
}

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

func FromManifest(man *manifest.ProjectManifest) (Provider, error) {
	switch man.Provider {
	case api.ProviderGCP:
		return gcpFromManifest(man)
	case api.ProviderAWS:
		return awsFromManifest(man)
	case api.ProviderAzure:
		return AzureFromManifest(man, nil)
	case api.ProviderEquinix:
		return equinixFromManifest(man)
	case api.ProviderKind:
		return kindFromManifest(man)
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
	case api.ProviderEquinix:
		return mkEquinix(conf)
	case api.ProviderKind:
		return mkKind(conf)
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
