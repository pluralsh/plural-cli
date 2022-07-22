package api

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/AlecAivazis/survey/v2"
	"gopkg.in/yaml.v2"

	"github.com/pluralsh/gqlclient"
	"github.com/pluralsh/gqlclient/pkg/utils"
	fileutils "github.com/pluralsh/plural/pkg/utils"
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
	Private       bool   `json:"private" yaml:"private,omitempty"`
	Tags          []Tag  `json:"tags,omitempty" yaml:"tags"`
	Icon          string `json:"icon,omitempty" yaml:"icon"`
	DarkIcon      string `json:"darkIcon,omitempty" yaml:"darkIcon"`
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

	return &Repository{
		Id:          resp.Repository.ID,
		Name:        resp.Repository.Name,
		Description: utils.ConvertStringPointer(resp.Repository.Description),
		Icon:        utils.ConvertStringPointer(resp.Repository.Icon),
		DarkIcon:    utils.ConvertStringPointer(resp.Repository.DarkIcon),
		Notes:       utils.ConvertStringPointer(resp.Repository.Notes),
		Publisher: &Publisher{
			Name: resp.Repository.Publisher.Name,
		},
	}, nil

}

func (client *client) CreateRepository(name, publisher string, input *gqlclient.RepositoryAttributes) error {
	var uploads []gqlclient.Upload

	iconUpload, err := getIconReader(input.Icon, "icon")
	if err != nil {
		return err
	}

	if iconUpload != nil {
		icon := "icon"
		input.Icon = &icon
		uploads = append(uploads, *iconUpload)
	}

	darkIconUpload, err := getIconReader(input.DarkIcon, "darkicon")
	if err != nil {
		return err
	}

	if darkIconUpload != nil {
		darkIcon := "darkicon"
		input.DarkIcon = &darkIcon
		uploads = append(uploads, *darkIconUpload)
	}

	if input.Notes != nil {
		file, _ := filepath.Abs(*input.Notes)
		notes, err := fileutils.ReadFile(file)
		if err != nil {
			return err
		}

		input.Notes = &notes
	}

	if _, err := client.pluralClient.CreateRepository(context.Background(), name, publisher, *input, gqlclient.WithFiles(uploads)); err != nil {
		return err
	}

	return nil
}

func (client *client) AcquireLock(repo string) (*ApplyLock, error) {
	resp, err := client.pluralClient.AcquireLock(client.ctx, repo)
	if err != nil {
		return nil, err
	}

	return &ApplyLock{
		Id:   resp.AcquireLock.ID,
		Lock: utils.ConvertStringPointer(resp.AcquireLock.Lock),
	}, err
}

func (client *client) ReleaseLock(repo, lock string) (*ApplyLock, error) {
	resp, err := client.pluralClient.ReleaseLock(client.ctx, repo, gqlclient.LockAttributes{Lock: lock})
	if err != nil {
		return nil, err
	}

	return &ApplyLock{
		Id:   resp.ReleaseLock.ID,
		Lock: utils.ConvertStringPointer(resp.ReleaseLock.Lock),
	}, nil
}

func (client *client) UnlockRepository(name string) error {
	_, err := client.pluralClient.UnlockRepository(client.ctx, name)
	if err != nil {
		return err
	}

	return nil
}

func (client *client) ListRepositories(query string) ([]*Repository, error) {
	resp, err := client.pluralClient.ListRepositories(client.ctx, &query)
	if err != nil {
		return nil, err
	}

	res := make([]*Repository, 0)
	for _, edge := range resp.Repositories.Edges {
		res = append(res, &Repository{
			Id:          edge.Node.ID,
			Name:        edge.Node.Name,
			Description: utils.ConvertStringPointer(edge.Node.Description),
			Icon:        utils.ConvertStringPointer(edge.Node.Icon),
			DarkIcon:    utils.ConvertStringPointer(edge.Node.DarkIcon),
			Notes:       utils.ConvertStringPointer(edge.Node.Notes),
			Publisher: &Publisher{
				Name: edge.Node.Publisher.Name,
			},
		})
	}

	return res, err
}

func (client *client) Scaffolds(in *ScaffoldInputs) ([]*ScaffoldFile, error) {

	scaffolds, err := client.pluralClient.Scaffolds(context.Background(), in.Application, in.Publisher, gqlclient.Category(strings.ToUpper(in.Category)), &in.Ingress, &in.Postgres)
	if err != nil {
		return nil, err
	}

	resp := make([]*ScaffoldFile, 0)

	for _, scaffold := range scaffolds.Scaffold {
		resp = append(resp, &ScaffoldFile{
			Path:    utils.ConvertStringPointer(scaffold.Path),
			Content: utils.ConvertStringPointer(scaffold.Content),
		})
	}

	return resp, err
}

func getIconReader(icon *string, field string) (*gqlclient.Upload, error) {
	if icon == nil {
		return nil, nil
	}

	file, err := filepath.Abs(*icon)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	return &gqlclient.Upload{
		Field: field,
		Name:  file,
		R:     f,
	}, nil
}

func ConstructRepositoryInput(marshalled []byte) (input *RepositoryInput, err error) {
	input = &RepositoryInput{}
	err = yaml.Unmarshal(marshalled, input)
	return
}

func ConstructGqlClientRepositoryInput(marshalled []byte) (*gqlclient.RepositoryAttributes, error) {
	input := &gqlclient.RepositoryAttributes{}
	if err := yaml.Unmarshal(marshalled, input); err != nil {
		return nil, err
	}
	return input, nil
}

func ConstructResourceDefinition(marshalled []byte) (input gqlclient.ResourceDefinitionAttributes, err error) {
	err = yaml.Unmarshal(marshalled, &input)
	return
}

func ConstructIntegration(marshalled []byte) (gqlclient.IntegrationAttributes, error) {
	intAttr := gqlclient.IntegrationAttributes{}
	err := yaml.Unmarshal(marshalled, &intAttr)
	if err != nil {
		return gqlclient.IntegrationAttributes{}, err
	}
	return intAttr, nil
}
