package api

import (
	"fmt"
	"os"
	"strings"
	"path/filepath"

	"gopkg.in/yaml.v2"
	"github.com/michaeljguarino/graphql"
	"github.com/pluralsh/plural/pkg/utils"
	_ "github.com/AlecAivazis/survey/v2"
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

type OauthSettings struct {
	UriFormat  string `yaml:"uriFormat"`
	AuthMethod string `yaml:"authMethod"`
}

type RepositoryInput struct {
	Name          string
	Description   string
	Private       bool   `json:"private,omitempty" yaml:"private,omitempty"`
	Tags          []Tag  `json:"tags,omitempty" yaml:"tags"`
	Icon          string `json:"icon,omitempty" yaml:"icon"`
	DarkIcon      string `json:"darkIcon,omitempty" yaml:"darkIcon"`
	Category      string
	Notes         string `json:"notes,omitempty" yaml:"notes"`
	GitUrl        string `json:"gitUrl" yaml:"gitUrl"`
	Homepage      string `json:"homepage" yaml:"homepage"`
	OauthSettings *OauthSettings `yaml:"oauthSettings,omitempty"`
}

type LockAttributes struct {
	Lock string
}

type ScaffoldInputs struct {
	Application string `survey:"application"`
	Publisher   string `survey:"publisher"`
	Category    string `survey:"category"`
	Ingress     bool   `survey:"ingress"`
	Postgres    bool   `survey:"postgres"`
}

const updateRepository = `
	mutation UpdateRepository($name: String!, $input: ResourceDefinitionAttributes!) {
		updateRepository(repositoryName: $name, attributes: {integrationResourceDefinition: $input}) {
			id
		}
	}
`

const upsertRepository = `
	mutation UpsertRepository($name: String!, $publisher: String!, $attributes: RepositoryAttributes!) {
		upsertRepository(name: $name, publisher: $publisher, attributes: $attributes) { id }
	}
`

const createIntegration = `
	mutation CreateIntegration($name: String!, $attrs: IntegrationAttributes!) {
		createIntegration(repositoryName: $name, attributes: $attrs) { id }
	}
`

const updateRepo = `
	mutation UpdateRepo($name: String!, $attrs: RepositoryAttributes!) {
		updateRepository(repositoryName: $name, attributes: $attrs) { id }
	}
`

var getRepo = fmt.Sprintf(`
	query Repo($name: String) {
		repository(name: $name) { ...RepositoryFragment }
	}
	%s
`, RepositoryFragment)

var listRepos = fmt.Sprintf(`
	query Repos($q: String) {
		repositories(q: $q, first: 100) {
			edges { node { ...RepositoryFragment } }
		}
	}
	%s
`, RepositoryFragment)

var acquireLock = fmt.Sprintf(`
	mutation Acquire($name: String!) {
		acquireLock(repository: $name) { ...ApplyLockFragment }
	}
	%s
`, ApplyLockFragment)

var releaseLock = fmt.Sprintf(`
	mutation Acquire($name: String!, $attrs: LockAttributes!) {
		releaseLock(repository: $name, attributes: $attrs) { ...ApplyLockFragment	}
	}
	%s
`, ApplyLockFragment)

const scaffoldsQuery = `
	query Scaffolds($app: String!, $pub: String!, $cat: Category!, $ing: Boolean, $pg: Boolean) {
		scaffold(application: $app, publisher: $pub, category: $cat, ingress: $ing, postgres: $pg) {
			path
			content
		}
	}
`

const unlockRepository = `
	mutation Unlock($name: String!) {
		unlockRepository(name: $name)
	}
`

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

func (client *Client) UpdateRepository(name string, input *RepositoryInput) (string, error) {
	var resp struct {
		Id string
	}
	req := client.Build(updateRepo)
	req.Var("attrs", input)
	req.Var("name", name)
	err := client.Run(req, &resp)
	return resp.Id, err
}

func (client *Client) CreateRepository(name, publisher string, input *RepositoryInput) error {
	var resp struct {
		UpsertRepository struct {
			Id string
		}
	}
	
	req := client.Build(upsertRepository)
	req.Var("name", name)
	req.Var("publisher", publisher)

	ok, err := getIconReader(input.Icon, "icon", req)
	if err != nil {
		return err
	}

	if ok {
		input.Icon = "icon"
	}

	ok, err = getIconReader(input.DarkIcon, "darkicon", req)
	if err != nil {
		return err
	}

	if ok {
		input.DarkIcon = "darkicon"
	}
	
	if input.Notes != "" {
		file, _ := filepath.Abs(input.Notes)
		notes, err := utils.ReadFile(file)
		if err != nil {
			return err
		}

		input.Notes = notes
	}

	req.Var("attributes", input)
	return client.Run(req, &resp)
}

func (client *Client) AcquireLock(repo string) (*ApplyLock, error) {
	var resp struct {
		AcquireLock *ApplyLock
	}

	req := client.Build(acquireLock)
	req.Var("name", repo)
	err := client.Run(req, &resp)
	return resp.AcquireLock, err
}

func (client *Client) ReleaseLock(repo, lock string) (*ApplyLock, error) {
	var resp struct {
		ReleaseLock *ApplyLock
	}

	req := client.Build(releaseLock)
	req.Var("name", repo)
	req.Var("attrs", LockAttributes{Lock: lock})
	err := client.Run(req, &resp)
	return resp.ReleaseLock, err
}

func (client *Client) UnlockRepository(name string) error {
	var resp struct {
		UnlockRepository int
	}

	req := client.Build(unlockRepository)
	req.Var("name", name)
	return client.Run(req, &resp)
}

func (client *Client) ListRepositories(query string) ([]*Repository, error) {
	var resp struct {
		Repositories struct {
			Edges []*RepositoryEdge
		}
	}

	req := client.Build(listRepos)
	req.Var("q", query)
	err := client.Run(req, &resp)
	res := make([]*Repository, len(resp.Repositories.Edges))
	for i, edge := range resp.Repositories.Edges {
		res[i] = edge.Node
	}

	return res, err
}

func (client *Client) Scaffolds(in *ScaffoldInputs) ([]*ScaffoldFile, error) {
	var resp struct {
		Scaffold []*ScaffoldFile
	}

	req := client.Build(scaffoldsQuery)
	req.Var("app", in.Application)
	req.Var("pub", in.Publisher)
	req.Var("cat", strings.ToUpper(in.Category))
	req.Var("ing", in.Ingress)
	req.Var("pg", in.Postgres)
	err := client.Run(req, &resp)
	return resp.Scaffold, err
}

func getIconReader(icon, field string, req *graphql.Request) (bool, error) {
	if icon == "" {
		return false, nil
	}

	file, err := filepath.Abs(icon)
	if err != nil {
		return false, err
	}
	f, err := os.Open(file)
	req.File(field, file, f)
	return true, err
}

func ConstructRepositoryInput(marshalled []byte) (input *RepositoryInput, err error) {
	input = &RepositoryInput{}
	err = yaml.Unmarshal(marshalled, input)
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
