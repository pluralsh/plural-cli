package api

import (
	"context"
	"net/http"

	"github.com/pluralsh/gqlclient"
	"github.com/pluralsh/plural/pkg/config"
)

type authedTransport struct {
	key     string
	wrapped http.RoundTripper
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.key)
	return t.wrapped.RoundTrip(req)
}

type Client interface {
	ListArtifacts(repo string) ([]Artifact, error)
	CreateArtifact(repo string, attrs ArtifactAttributes) (Artifact, error)
	Me() (*Me, error)
	LoginMethod(email string) (*LoginMethod, error)
	PollLoginToken(token string) (string, error)
	DeviceLogin() (*DeviceLogin, error)
	Login(email string, pwd string) (string, error)
	ImpersonateServiceAccount(email string) (string, string, error)
	CreateAccessToken() (string, error)
	GrabAccessToken() (string, error)
	ListKeys(emails []string) ([]*PublicKey, error)
	CreateKey(name string, content string) error
	GetEabCredential(cluster string, provider string) (*EabCredential, error)
	DeleteEabCredential(cluster string, provider string) error
	CreateEvent(event *UserEventAttributes) error
	GetTfProviders() ([]string, error)
	GetTfProviderScaffold(name string, version string) (string, error)
	GetRepository(repo string) (*Repository, error)
	CreateRepository(name string, publisher string, input *gqlclient.RepositoryAttributes) error
	AcquireLock(repo string) (*ApplyLock, error)
	ReleaseLock(repo string, lock string) (*ApplyLock, error)
	UnlockRepository(name string) error
	ListRepositories(query string) ([]*Repository, error)
	Scaffolds(in *ScaffoldInputs) ([]*ScaffoldFile, error)
	UpdateVersion(spec *VersionSpec, tags []string) error
	GetCharts(repoId string) ([]*Chart, error)
	GetVersions(chartId string) ([]*Version, error)
	GetChartInstallations(repoId string) ([]*ChartInstallation, error)
	GetPackageInstallations(repoId string) (charts []*ChartInstallation, tfs []*TerraformInstallation, err error)
	CreateCrd(repo string, chart string, file string) error
	CreateDomain(name string) error
	GetInstallation(name string) (*Installation, error)
	GetInstallationById(id string) (*Installation, error)
	GetInstallations() ([]*Installation, error)
	OIDCProvider(id string, attributes *OidcProviderAttributes) error
	ResetInstallations() (int, error)
	CreateRecipe(repoName string, attrs gqlclient.RecipeAttributes) (string, error)
	GetRecipe(repo string, name string) (*Recipe, error)
	ListRecipes(repo string, provider string) ([]*Recipe, error)
	InstallRecipe(id string) error
	GetShell() (CloudShell, error)
	DeleteShell() error
	GetTerraforma(repoId string) ([]*Terraform, error)
	GetTerraformInstallations(repoId string) ([]*TerraformInstallation, error)
	UploadTerraform(dir string, repoName string) (Terraform, error)
	GetStack(name, provider string) (*Stack, error)
	CreateStack(attributes gqlclient.StackAttributes) (string, error)
	ListStacks(featured bool) ([]*Stack, error)
	UninstallChart(id string) error
	UninstallTerraform(id string) error
}

type client struct {
	ctx          context.Context
	pluralClient *gqlclient.Client
	config       config.Config
}

func NewClient() Client {
	conf := config.Read()
	return FromConfig(&conf)
}

func FromConfig(conf *config.Config) Client {
	httpClient := http.Client{
		Transport: &authedTransport{
			key:     conf.Token,
			wrapped: http.DefaultTransport,
		},
	}

	return &client{
		pluralClient: gqlclient.NewClient(&httpClient, conf.Url()),
		config:       *conf,
		ctx:          context.Background(),
	}

}
