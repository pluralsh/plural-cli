package api

import (
	"context"
	"encoding/json"
	"net/http"

	rawclient "github.com/Yamashou/gqlgenc/clientv2"
	"github.com/pkg/errors"
	"github.com/pluralsh/gqlclient"

	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/utils"
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
	CreateCrd(repo string, chart string, file string) error
	CreateDomain(name string) error
	CreateInstallation(id string) (string, error)
	GetInstallation(name string) (*Installation, error)
	GetInstallations() ([]*Installation, error)
	OIDCProvider(id string, attributes *OidcProviderAttributes) error
	GetShell() (CloudShell, error)
	DeleteShell() error
	GetTerraform(repoId string) ([]*Terraform, error)
	CreateKeyBackup(attrs KeyBackupAttributes) error
	GetKeyBackup(name string) (*KeyBackup, error)
	ListKeyBackups() ([]*KeyBackup, error)
	GetHelp(prompt string) (string, error)
	DestroyCluster(domain, name, provider string) error
	CreateDependency(source, dest string) error
	PromoteCluster() error
	Clusters() ([]*Cluster, error)
	Cluster(id string) (*Cluster, error)
	CreateUpgrade(queue, repository string, attrs gqlclient.UpgradeAttributes) error
	TransferOwnership(name, email string) error
	Release(name string, tags []string) error
	Chat(history []*ChatMessage) (*ChatMessage, error)
	InstallVersion(tp, repo, pkg, vsn string) error
	CreateTrust(issuer, trust string) error
	DeleteTrust(id string) error
	OidcToken(provider gqlclient.ExternalOidcProvider, token, email string) (string, error)
	MarkSynced(repo string) error
	GetConsoleInstances() ([]*gqlclient.ConsoleInstanceFragment, error)
	UpdateConsoleInstance(id string, attrs gqlclient.ConsoleInstanceUpdateAttributes) error
}

type client struct {
	ctx          context.Context
	pluralClient *gqlclient.Client
	config       config.Config
	httpClient   *http.Client
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
		pluralClient: gqlclient.NewClient(&httpClient, conf.Url(), nil),
		config:       *conf,
		ctx:          context.Background(),
		httpClient:   &httpClient,
	}

}

func GetErrorResponse(err error, methodName string) error {
	if err == nil {
		return nil
	}
	utils.LogError().Println(err)
	errResponse := &rawclient.ErrorResponse{}
	newErr := json.Unmarshal([]byte(err.Error()), errResponse)
	if newErr != nil {
		return err
	}

	errList := errors.New(methodName)
	if errResponse.GqlErrors != nil {
		for _, err := range *errResponse.GqlErrors {
			errList = errors.Wrap(errList, err.Message)
		}
		errList = errors.Wrap(errList, "GraphQL error")
	}
	if errResponse.NetworkError != nil {
		errList = errors.Wrap(errList, errResponse.NetworkError.Message)
		errList = errors.Wrap(errList, "Network error")
	}

	return errList
}
