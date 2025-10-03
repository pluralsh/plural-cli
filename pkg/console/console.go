package console

import (
	"context"
	"fmt"
	"net/http"
	neturl "net/url"

	consoleclient "github.com/pluralsh/console/go/client"
)

type consoleClient struct {
	ctx    context.Context
	client consoleclient.ConsoleClient
	url    string
	extUrl string
	token  string
}

type ConsoleClient interface {
	Url() string
	ExtUrl() string
	Token() string
	AgentUrl(id string) (string, error)
	ListClusters() (*consoleclient.ListClusters, error)
	GetProject(name string) (*consoleclient.ProjectFragment, error)
	GetCluster(clusterId, clusterName *string) (*consoleclient.ClusterFragment, error)
	GetDeployToken(clusterId, clusterName *string) (string, error)
	UpdateCluster(id string, attr consoleclient.ClusterUpdateAttributes) (*consoleclient.UpdateCluster, error)
	DeleteCluster(id string) error
	DetachCluster(id string) error
	ListClusterServices(clusterId, handle *string) ([]*consoleclient.ServiceDeploymentEdgeFragment, error)
	CreateRepository(url string, privateKey, passphrase, username, password *string) (*consoleclient.CreateGitRepository, error)
	GetRepository(id string) (*consoleclient.GetGitRepository, error)
	ListRepositories() (*consoleclient.ListGitRepositories, error)
	UpdateRepository(id string, attrs consoleclient.GitAttributes) (*consoleclient.UpdateGitRepository, error)
	CreateClusterService(clusterId, clusterName *string, attr consoleclient.ServiceDeploymentAttributes) (*consoleclient.ServiceDeploymentExtended, error)
	UpdateClusterService(serviceId, serviceName, clusterName *string, attributes consoleclient.ServiceUpdateAttributes) (*consoleclient.ServiceDeploymentExtended, error)
	CloneService(clusterId string, serviceId, serviceName, clusterName *string, attributes consoleclient.ServiceCloneAttributes) (*consoleclient.ServiceDeploymentFragment, error)
	GetClusterService(serviceId, serviceName, clusterName *string) (*consoleclient.ServiceDeploymentExtended, error)
	DeleteClusterService(serviceId string) (*consoleclient.DeleteServiceDeployment, error)
	ListProviders() (*consoleclient.ListProviders, error)
	CreateProviderCredentials(name string, attr consoleclient.ProviderCredentialAttributes) (*consoleclient.CreateProviderCredential, error)
	DeleteProviderCredentials(id string) (*consoleclient.DeleteProviderCredential, error)
	SavePipeline(name string, attrs consoleclient.PipelineAttributes) (*consoleclient.PipelineFragmentMinimal, error)
	CreatePipelineContext(id string, attrs consoleclient.PipelineContextAttributes) (*consoleclient.PipelineContextFragment, error)
	GetPipelineContext(id string) (*consoleclient.PipelineContextFragment, error)
	CreateCluster(attributes consoleclient.ClusterAttributes) (*consoleclient.CreateCluster, error)
	CreateProvider(attr consoleclient.ClusterProviderAttributes) (*consoleclient.CreateClusterProvider, error)
	MyCluster() (*consoleclient.MyCluster, error)
	SaveServiceContext(name string, attributes consoleclient.ServiceContextAttributes) (*consoleclient.ServiceContextFragment, error)
	GetServiceContext(name string) (*consoleclient.ServiceContextFragment, error)
	KickClusterService(serviceId, serviceName, clusterName *string) (*consoleclient.ServiceDeploymentExtended, error)
	ListNotificationSinks(after *string, first *int64) (*consoleclient.ListNotificationSinks_NotificationSinks, error)
	CreateNotificationSinks(attr consoleclient.NotificationSinkAttributes) (*consoleclient.NotificationSinkFragment, error)
	UpdateDeploymentSettings(attr consoleclient.DeploymentSettingsAttributes) (*consoleclient.UpdateDeploymentSettings, error)
	GetGlobalSettings() (*consoleclient.DeploymentSettingsFragment, error)
	ListStackRuns(stackID string) (*consoleclient.ListStackRuns, error)
	CreatePullRequest(id string, branch, context *string) (*consoleclient.PullRequestFragment, error)
	GetPrAutomationByName(name string) (*consoleclient.PrAutomationFragment, error)
	CreateBootstrapToken(attributes consoleclient.BootstrapTokenAttributes) (string, error)
	CreateClusterRegistration(attributes consoleclient.ClusterRegistrationCreateAttributes) (*consoleclient.ClusterRegistrationFragment, error)
	IsClusterRegistrationComplete(machineID string) (bool, *consoleclient.ClusterRegistrationFragment)
	GetUser(email string) (*consoleclient.UserFragment, error)
}

type authedTransport struct {
	token   string
	wrapped http.RoundTripper
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Token "+t.token)
	return t.wrapped.RoundTrip(req)
}

func NewConsoleClient(token, url string) (ConsoleClient, error) {
	httpClient := http.Client{
		Transport: &authedTransport{
			token:   token,
			wrapped: http.DefaultTransport,
		},
	}

	return &consoleClient{
		url:    NormalizeUrl(url),
		extUrl: NormalizeExtUrl(url),
		token:  token,
		client: consoleclient.NewClient(&httpClient, NormalizeUrl(url), nil, consoleclient.PersistedQueryInterceptor),
		ctx:    context.Background(),
	}, nil
}

func NormalizeExtUrl(url string) string {
	parsed, err := neturl.Parse(url)
	if err != nil {
		panic(err)
	}

	// Trying to parse a hostname and path without a scheme is invalid but may not necessarily return an error,
	// due to parsing ambiguities.
	// The following block was added to support URLs without scheme set as we change it to HTTPS anyway.
	if parsed.Scheme == "" && parsed.Host == "" {
		if parsed, err = neturl.Parse("//" + url); err != nil {
			panic(err)
		}
	}

	return fmt.Sprintf("https://%s/ext/gql", parsed.Host)
}

func NormalizeUrl(url string) string {
	parsed, err := neturl.Parse(url)
	if err != nil {
		panic(err)
	}

	// Trying to parse a hostname and path without a scheme is invalid but may not necessarily return an error,
	// due to parsing ambiguities.
	// The following block was added to support URLs without scheme set as we change it to HTTPS anyway.
	if parsed.Scheme == "" && parsed.Host == "" {
		if parsed, err = neturl.Parse("//" + url); err != nil {
			panic(err)
		}
	}

	return fmt.Sprintf("https://%s/gql", parsed.Host)
}

func (c *consoleClient) Url() string {
	return c.url
}

func (c *consoleClient) ExtUrl() string {
	return c.extUrl
}

func (c *consoleClient) Token() string {
	return c.token
}
