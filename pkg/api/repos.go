package api

import (
	_ "github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/gqlclient"
	"github.com/pluralsh/gqlclient/pkg/utils"
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
	ReleaseStatus string `json:"releaseStatus,omitempty" yaml:"releaseStatus,omitempty"`
	Private       bool   `json:"private" yaml:"private,omitempty"`
	Tags          []Tag  `json:"tags,omitempty" yaml:"tags"`
	Icon          string `json:"icon,omitempty" yaml:"icon"`
	DarkIcon      string `json:"darkIcon,omitempty" yaml:"darkIcon"`
	Docs          string `json:"docs,omitempty" yaml:"docs"`
	Contributors  []string
	Category      string
	Notes         string         `json:"notes,omitempty" yaml:"notes"`
	GitUrl        string         `json:"gitUrl" yaml:"gitUrl"`
	Homepage      string         `json:"homepage" yaml:"homepage"`
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

func (client *client) GetRepository(repo string) (*Repository, error) {
	resp, err := client.pluralClient.GetRepository(client.ctx, &repo)
	if err != nil {
		return nil, err
	}

	return convertRepository(resp.Repository), nil
}

func convertRepository(repo *gqlclient.RepositoryFragment) *Repository {
	return &Repository{
		Id:          repo.ID,
		Name:        repo.Name,
		Description: utils.ConvertStringPointer(repo.Description),
		Icon:        utils.ConvertStringPointer(repo.Icon),
		DarkIcon:    utils.ConvertStringPointer(repo.DarkIcon),
		Notes:       utils.ConvertStringPointer(repo.Notes),
		Publisher: &Publisher{
			Name: repo.Publisher.Name,
		},
	}
}
