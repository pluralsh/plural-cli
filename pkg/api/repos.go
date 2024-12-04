package api

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/AlecAivazis/survey/v2"
	"sigs.k8s.io/yaml"

	"github.com/pluralsh/gqlclient"
	"github.com/pluralsh/gqlclient/pkg/utils"
	fileutils "github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/samber/lo"
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

func (client *client) CreateRepository(name, publisher string, input *gqlclient.RepositoryAttributes) error {
	var uploads []gqlclient.Upload

	iconUpload, err := getIconReader(input.Icon, "icon")
	if err != nil {
		return err
	}

	if iconUpload != nil {
		input.Icon = lo.ToPtr("icon")
		uploads = append(uploads, *iconUpload)
	}

	darkIconUpload, err := getIconReader(input.DarkIcon, "darkicon")
	if err != nil {
		return err
	}

	if darkIconUpload != nil {
		input.DarkIcon = lo.ToPtr("darkicon")
		uploads = append(uploads, *darkIconUpload)
	}

	if input.Docs != nil && *input.Docs != "" {
		tarFile, err := tarDir(name, *input.Docs, "")
		if err != nil {
			return err
		}
		defer os.Remove(tarFile)

		docsUpload, err := getIconReader(lo.ToPtr(tarFile), "docs")
		if err != nil {
			return err
		}
		input.Docs = lo.ToPtr("docs")
		uploads = append(uploads, *docsUpload)
	}

	if input.Notes != nil && *input.Notes != "" {
		file, _ := filepath.Abs(*input.Notes)
		notes, err := fileutils.ReadFile(file)
		if err != nil {
			return err
		}

		input.Notes = &notes
	}
	client.pluralClient.Client.CustomDo = gqlclient.WithFiles(uploads, client.httpClient)
	_, err = client.pluralClient.CreateRepository(context.Background(), name, publisher, *input)
	return err
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

func (client *client) ListRepositories(query string) ([]*Repository, error) {
	resp, err := client.pluralClient.ListRepositories(client.ctx, &query)
	if err != nil {
		return nil, err
	}

	res := make([]*Repository, 0)
	for _, edge := range resp.Repositories.Edges {
		rep := &Repository{
			Id:          edge.Node.ID,
			Name:        edge.Node.Name,
			Description: utils.ConvertStringPointer(edge.Node.Description),
			Icon:        utils.ConvertStringPointer(edge.Node.Icon),
			DarkIcon:    utils.ConvertStringPointer(edge.Node.DarkIcon),
			Notes:       utils.ConvertStringPointer(edge.Node.Notes),
			Publisher: &Publisher{
				Name: edge.Node.Publisher.Name,
			},
			Recipes: []*Recipe{},
		}
		for _, rcp := range edge.Node.Recipes {
			rep.Recipes = append(rep.Recipes, &Recipe{Name: rcp.Name})
		}
		res = append(res, rep)
	}

	return res, err
}

func (client *client) Release(name string, tags []string) error {
	_, err := client.pluralClient.Release(context.Background(), name, tags)
	return err
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
	if *icon == "" {
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
	repoInput, err := ConstructRepositoryInput(marshalled)
	if err != nil {
		return nil, err
	}

	category := gqlclient.Category(repoInput.Category)

	var releaseStatus *gqlclient.ReleaseStatus
	if repoInput.ReleaseStatus != "" {
		releaseStatus = lo.ToPtr(gqlclient.ReleaseStatus(repoInput.ReleaseStatus))
	}

	resp := &gqlclient.RepositoryAttributes{
		Category:      &category,
		DarkIcon:      &repoInput.DarkIcon,
		Description:   &repoInput.Description,
		ReleaseStatus: releaseStatus,
		Contributors:  lo.ToSlicePtr(repoInput.Contributors),
		GitURL:        &repoInput.GitUrl,
		Homepage:      &repoInput.Homepage,
		Icon:          &repoInput.Icon,
		Docs:          &repoInput.Docs,
		Name:          &repoInput.Name,
		Notes:         &repoInput.Notes,
		Private:       &repoInput.Private,
		Tags:          []*gqlclient.TagAttributes{},
	}
	if repoInput.OauthSettings != nil {
		resp.OauthSettings = &gqlclient.OauthSettingsAttributes{
			AuthMethod: gqlclient.OidcAuthMethod(repoInput.OauthSettings.AuthMethod),
			URIFormat:  repoInput.OauthSettings.UriFormat,
		}
	}
	for _, tag := range repoInput.Tags {
		resp.Tags = append(resp.Tags, &gqlclient.TagAttributes{
			Tag: tag.Tag,
		})
	}

	return resp, nil
}
