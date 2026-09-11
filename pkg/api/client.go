package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	rawclient "github.com/Yamashou/gqlgenc/clientv2"
	"github.com/pluralsh/gqlclient"

	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/utils"
	clierrors "github.com/pluralsh/plural-cli/pkg/utils/errors"
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
	GetRepository(repo string) (*Repository, error)
	UpdateVersion(spec *VersionSpec, tags []string) error
	CreateDomain(name string) error
	CreateInstallation(id string) (string, error)
	GetInstallation(name string) (*Installation, error)
	OIDCProvider(id string, attributes *OidcProviderAttributes) error
	CreateKeyBackup(attrs KeyBackupAttributes) error
	GetKeyBackup(name string) (*KeyBackup, error)
	ListKeyBackups() ([]*KeyBackup, error)
	GetHelp(prompt string) (string, error)
	Clusters() ([]*Cluster, error)
	Cluster(id string) (*Cluster, error)
	Chat(history []*ChatMessage) (*ChatMessage, error)
	CreateTrust(issuer, trust string) error
	DeleteTrust(id string) error
	OidcToken(provider gqlclient.ExternalOidcProvider, token, email string) (string, error)
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
		pluralClient: gqlclient.NewClient(&httpClient, conf.Url(), nil, clierrors.GraphQLInterceptor),
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
	var errResponse *rawclient.ErrorResponse
	if !errors.As(err, &errResponse) {
		return err
	}
	return fmt.Errorf("%s: %w", methodName, clierrors.GraphQL(err))
}
