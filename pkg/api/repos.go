package api

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type ResourceDefinitionInput struct {
	Name string
	Spec []Specification
}

type Specification struct {
	Name     string
	Type     string
	Inner    string `json:"inner,omitempty"`
	Required bool
	Spec     []Specification `json:"spec,omitempty"`
}

type IntegrationInput struct {
	Name        string
	Description string
	Icon        string
	SourceURL   string `json:"sourceUrl,omitempty"`
	Spec        string
	Type        string `json:"type,omitempty"`
	Tags        []Tag  `json:"tags,omitempty" yaml:"tags"`
}

type RepositoryInput struct {
	Dashboards []struct {
		Name string
		UID  string `json:"uid"`
	}
}

const updateRepository = `
	mutation UpdateRepository($name: String!, $input: ResourceDefinitionAttributes!) {
		updateRepository(repositoryName: $name, attributes: {integrationResourceDefinition: $input}) {
			id
		}
	}
`

const createIntegration = `
	mutation CreateIntegration($name: String!, $attrs: IntegrationAttributes!) {
		createIntegration(repositoryName: $name, attributes: $attrs) {
			id
		}
	}
`

const updateRepo = `
	mutation UpdateRepo($name: String!, $attrs: RepositoryAttributes!) {
		updateRepository(repositoryName: $name, attributes: $attrs) {
			id
		}
	}
`

var getRepo = fmt.Sprintf(`
	query Repo($name: String) {
		repository(name: $name) {
			...RepositoryFragment
		}
	}
	%s
`, RepositoryFragment)

func (client *Client) GetRepository(repo string) (repository *Repository, err error) {
	var resp struct {
		Repository *Repository
	}
	req := client.Build(getRepo)
	req.Var("name", repo)
	err = client.Run(req, &resp)
	repository = resp.Repository
	return
}

func (client *Client) CreateResourceDefinition(repoName string, input ResourceDefinitionInput) (string, error) {
	var resp struct {
		Id string
	}
	req := client.Build(updateRepository)
	req.Var("input", input)
	req.Var("name", repoName)
	err := client.Run(req, &resp)
	return resp.Id, err
}

func (client *Client) CreateIntegration(name string, input IntegrationInput) (string, error) {
	var resp struct {
		Id string
	}
	req := client.Build(createIntegration)
	req.Var("attrs", input)
	req.Var("name", name)
	err := client.Run(req, &resp)
	return resp.Id, err
}

func (client *Client) UpdateRepository(name string, input RepositoryInput) (string, error) {
	var resp struct {
		Id string
	}
	req := client.Build(updateRepo)
	req.Var("attrs", input)
	req.Var("name", name)
	err := client.Run(req, &resp)
	return resp.Id, err
}

func ConstructRepositoryInput(marshalled []byte) (input RepositoryInput, err error) {
	err = yaml.Unmarshal(marshalled, &input)
	return
}

func ConstructResourceDefinition(marshalled []byte) (input ResourceDefinitionInput, err error) {
	err = yaml.Unmarshal(marshalled, &input)
	return
}

func ConstructIntegration(marshalled []byte) (IntegrationInput, error) {
	var intg struct {
		Name        string
		Description string
		Icon        string
		SourceURL   string `yaml:"sourceUrl"`
		Type        string
		Tags        []Tag
		Spec        interface{}
	}
	err := yaml.Unmarshal(marshalled, &intg)
	if err != nil {
		return IntegrationInput{}, err
	}

	str, err := yaml.Marshal(intg.Spec)
	return IntegrationInput{
		Name:        intg.Name,
		Description: intg.Description,
		Icon:        intg.Icon,
		Spec:        string(str),
		Tags:        intg.Tags,
		Type:        intg.Type,
		SourceURL:   intg.SourceURL,
	}, err
}
