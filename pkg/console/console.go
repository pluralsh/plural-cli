package console

import (
	"context"
	"fmt"
	"net/http"

	consoleclient "github.com/pluralsh/console-client-go"
)

type consoleClient struct {
	ctx    context.Context
	client consoleclient.ConsoleClient
	url    string
	token  string
}

type ConsoleClient interface {
	Url() string
	Token() string
	ListClusters() (*consoleclient.ListClusters, error)
	GetCluster(clusterId, clusterName *string) (*consoleclient.ClusterFragment, error)
	GetDeployToken(clusterId, clusterName *string) (string, error)
	UpdateCluster(id string, attr consoleclient.ClusterUpdateAttributes) (*consoleclient.UpdateCluster, error)
	DeleteCluster(id string) error
	DetachCluster(id string) error
	ListClusterServices(clusterId, handle *string) ([]*consoleclient.ServiceDeploymentEdgeFragment, error)
	CreateRepository(url string, privateKey, passphrase, username, password *string) (*consoleclient.CreateGitRepository, error)
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
	SavePipeline(name string, attrs consoleclient.PipelineAttributes) (*consoleclient.PipelineFragment, error)
	CreateCluster(attributes consoleclient.ClusterAttributes) (*consoleclient.CreateCluster, error)
	CreateProvider(attr consoleclient.ClusterProviderAttributes) (*consoleclient.CreateClusterProvider, error)
	MyCluster() (*consoleclient.MyCluster, error)
	SaveServiceContext(name string, attributes consoleclient.ServiceContextAttributes) (*consoleclient.ServiceContextFragment, error)
	GetServiceContext(name string) (*consoleclient.ServiceContextFragment, error)
	KickClusterService(serviceId, serviceName, clusterName *string) (*consoleclient.ServiceDeploymentExtended, error)
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
		url:    url,
		token:  token,
		client: consoleclient.NewClient(&httpClient, fmt.Sprintf("%s/gql", url), nil),
		ctx:    context.Background(),
	}, nil
}

func (c *consoleClient) Url() string {
	return c.url
}

func (c *consoleClient) Token() string {
	return c.token
}
