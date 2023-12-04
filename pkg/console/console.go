package console

import (
	"context"
	"fmt"
	"net/http"

	consoleclient "github.com/pluralsh/console-client-go"
)

type consoleClient struct {
	ctx    context.Context
	client *consoleclient.Client
	url    string
	token  string
}

type ConsoleClient interface {
	Url() string
	Token() string
	ListClusters() (*consoleclient.ListClusters, error)
	GetCluster(clusterId, clusterName *string) (*consoleclient.ClusterFragment, error)
	UpdateCluster(id string, attr consoleclient.ClusterUpdateAttributes) (*consoleclient.UpdateCluster, error)
	DeleteCluster(id string) error
	DetachCluster(id string) error
	ListClusterServices(clusterId, handle *string) ([]*consoleclient.ServiceDeploymentEdgeFragment, error)
	CreateRepository(url string, privateKey, passphrase, username, password *string) (*consoleclient.CreateGitRepository, error)
	ListRepositories() (*consoleclient.ListGitRepositories, error)
	UpdateRepository(id string, attrs consoleclient.GitAttributes) (*consoleclient.UpdateGitRepository, error)
	CreateClusterService(clusterId, clusterName *string, attr consoleclient.ServiceDeploymentAttributes) (*consoleclient.ServiceDeploymentFragment, error)
	UpdateClusterService(serviceId, serviceName, clusterName *string, attributes consoleclient.ServiceUpdateAttributes) (*consoleclient.ServiceDeploymentFragment, error)
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
}

func NewConsoleClient(token, url string) (ConsoleClient, error) {
	return &consoleClient{
		url:   url,
		token: token,
		client: consoleclient.NewClient(http.DefaultClient, fmt.Sprintf("%s/gql", url), func(req *http.Request) {
			req.Header.Set("Authorization", fmt.Sprintf("Token %s", token))
		}),
		ctx: context.Background(),
	}, nil
}

func (c *consoleClient) Url() string {
	return c.url
}

func (c *consoleClient) Token() string {
	return c.token
}
