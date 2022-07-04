package provider

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

type Provider interface {
	Name() string
	Cluster() string
	Project() string
	Region() string
	Bucket() string
	KubeConfig() error
	CreateBackend(prefix string, ctx map[string]interface{}) (string, error)
	Context() map[string]interface{}
	Decommision(node *v1.Node) error
	Preflights() []*Preflight
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

var providers = Providers{}

func GetProviderScaffold(provider string) (string, error) {
	if providers.Scaffolds == nil {
		providers.Scaffolds = make(map[string]string)
	}
	_, ok := providers.Scaffolds[provider]
	if !ok {
		client := api.NewClient()
		scaffold, err := client.GetTfProviderScaffold(provider)
		providers.Scaffolds[provider] = scaffold
		if err != nil {
			return "", err
		}
	}
	return providers.Scaffolds[provider], nil
}

func GetProvider() (Provider, error) {
	path := manifest.ProjectManifestPath()
	if project, err := manifest.ReadProject(path); err == nil {
		return FromManifest(project)
	}
	err := getAvailableProviders()
	if err != nil {
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
		return azureFromManifest(man)
	case EQUINIX:
		return equinixFromManifest(man)
	case KIND:
		return kindFromManifest(man)
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
		for i := range available {
			if available[i] == "GCP" {
				available[i] = "google"
			}
			available[i] = strings.ToLower(available[i])
		}
		if err != nil {
			return err
		}
		providers.AvailableProviders = available
	}
	return nil
}
