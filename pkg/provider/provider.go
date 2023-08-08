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
	utils.Highlight("Executing preflight check :: %s\n", pf.Name)
	if err := pf.Callback(); err != nil {
		fmt.Println("\nFound error:")
		return err
	}

	utils.Success("%s \u2713\n", pf.Name)
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
	case GCP:
		return gcpFromManifest(man)
	case AWS:
		return awsFromManifest(man)
	case AZURE:
		return AzureFromManifest(man, nil)
	case EQUINIX:
		return equinixFromManifest(man)
	case KIND:
		return kindFromManifest(man)
	case TEST:
		return testFromManifest(man)
	default:
		return nil, fmt.Errorf("invalid provider name: %s", man.Provider)
	}
}

func New(provider string) (Provider, error) {
	conf := config.Read()
	switch provider {
	case GCP:
		return mkGCP(conf)
	case AWS:
		return mkAWS(conf)
	case AZURE:
		return mkAzure(conf)
	case EQUINIX:
		return mkEquinix(conf)
	case KIND:
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
		providers.AvailableProviders = algorithms.Map(available, func(p string) string {
			if p == "GCP" {
				return "google"
			}
			return strings.ToLower(p)
		})
	}
	return nil
}
