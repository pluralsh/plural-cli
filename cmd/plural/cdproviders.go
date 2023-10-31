package plural

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pkg/errors"
	gqlclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/urfave/cli"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const kindSecret = "Secret"

func (p *Plural) handleDeleteProviderCredentials(c *cli.Context) error {
	id := c.Args().Get(0)
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	if _, err := p.ConsoleClient.DeleteProviderCredentials(id); err != nil {
		return err
	}
	utils.Success("Provider credential %s has been deleted successfully", id)
	return nil
}

func (p *Plural) handleCreateProviderCredentials(c *cli.Context) error {
	providerName := c.Args().Get(0)
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	attr, err := p.credentialsPreflights()
	if err != nil {
		return err
	}

	resp, err := p.ConsoleClient.CreateProviderCredentials(providerName, *attr)
	if err != nil {
		errList := errors.New("CreateProviderCredentials")
		errList = errors.Wrap(errList, err.Error())
		if *attr.Kind == kindSecret {
			if err := p.SecretDelete(*attr.Namespace, attr.Name); err != nil {
				errList = errors.Wrap(errList, err.Error())
			}
		}
		return errList
	}
	if resp == nil {
		return fmt.Errorf("the response from CreateProviderCredentials is empty")
	}

	headers := []string{"Id", "Name", "Namespace"}
	return utils.PrintTable([]*gqlclient.ProviderCredentialFragment{resp.CreateProviderCredential}, headers, func(sd *gqlclient.ProviderCredentialFragment) ([]string, error) {
		return []string{sd.ID, sd.Name, sd.Namespace}, nil
	})
}

func (p *Plural) handleListProviders(_ *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	providers, err := p.ConsoleClient.ListProviders()
	if err != nil {
		return err
	}
	if providers == nil {
		return fmt.Errorf("returned objects list [ListProviders] is nil")
	}

	headers := []string{"ID", "Name", "Cloud", "Editable", "Repo Url"}
	return utils.PrintTable(providers.ClusterProviders.Edges, headers, func(r *gqlclient.ClusterProviderEdgeFragment) ([]string, error) {
		editable := ""
		if r.Node.Editable != nil {
			editable = strconv.FormatBool(*r.Node.Editable)
		}
		repoUrl := ""
		if r.Node.Repository != nil {
			repoUrl = r.Node.Repository.URL
		}
		return []string{r.Node.ID, r.Node.Name, r.Node.Cloud, editable, repoUrl}, nil
	})
}

var availableProviders = []string{api.ProviderGCP, api.ProviderAzure, api.ProviderAWS}

func (p *Plural) credentialsPreflights() (*gqlclient.ProviderCredentialAttributes, error) {
	provider := ""
	prompt := &survey.Select{
		Message: "Select one of the following providers:",
		Options: availableProviders,
	}
	if err := survey.AskOne(prompt, &provider, survey.WithValidator(survey.Required)); err != nil {
		return nil, err
	}
	utils.Success("Using provider %s\n", provider)
	if provider == api.ProviderGCP {
		kind := kindSecret
		name, namespace, err := p.createSecret()
		if err != nil {
			return nil, err
		}
		return &gqlclient.ProviderCredentialAttributes{
			Namespace: &namespace,
			Name:      name,
			Kind:      &kind,
		}, nil
	}

	return nil, fmt.Errorf("unsupported provider")
}

func (p *Plural) createSecret() (name, namespace string, err error) {
	err = p.InitKube()
	if err != nil {
		return "", "", err
	}
	secretSurvey := []*survey.Question{
		{
			Name:     "name",
			Prompt:   &survey.Input{Message: "Enter the name of the secret: "},
			Validate: survey.Required,
		},
		{
			Name:     "namespace",
			Prompt:   &survey.Input{Message: "Enter the secret namespace: "},
			Validate: survey.Required,
		},
		{
			Name:     "data",
			Prompt:   &survey.Input{Message: "Enter the secret data pairs name=value, for example: user=admin password=abc : "},
			Validate: survey.Required,
		},
	}
	var resp struct {
		Name      string
		Namespace string
		Data      string
	}
	err = survey.Ask(secretSurvey, &resp)
	if err != nil {
		return
	}
	data := getSecretDataPairs(resp.Data)

	providerSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resp.Name,
			Namespace: resp.Namespace,
		},
		Data: data,
	}
	if _, err = p.SecretCreate(resp.Namespace, providerSecret); err != nil {
		return
	}
	name = resp.Name
	namespace = resp.Namespace
	return
}

func getSecretDataPairs(in string) map[string][]byte {
	res := map[string][]byte{}
	for _, conf := range strings.Split(in, " ") {
		configurationPair := strings.Split(conf, "=")
		if len(configurationPair) == 2 {
			res[configurationPair[0]] = []byte(configurationPair[1])
		}
	}
	return res
}
