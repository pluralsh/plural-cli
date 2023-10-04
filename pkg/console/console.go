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
}

type ConsoleClient interface {
	ListClusters() (*consoleclient.ListClusters, error)
	GetCluster(id string) (*consoleclient.GetCluster, error)
	UpdateCluster(id string, attr consoleclient.ClusterUpdateAttributes) (*consoleclient.UpdateCluster, error)
	ListClusterServices(clusterId, handle *string) ([]*consoleclient.ServiceDeploymentEdgeFragment, error)
	CreateRepository(url string, privateKey, passphrase, username, password *string) (*consoleclient.CreateGitRepository, error)
	ListRepositories() (*consoleclient.ListGitRepositories, error)
	UpdateRepository(id string, attrs consoleclient.GitAttributes) (*consoleclient.UpdateGitRepository, error)
	CreateClusterService(clusterId, clusterName *string, attr consoleclient.ServiceDeploymentAttributes) (*consoleclient.ServiceDeploymentFragment, error)
	UpdateClusterService(serviceId string, attr consoleclient.ServiceUpdateAttributes) (*consoleclient.UpdateServiceDeployment, error)
	GetClusterService(serviceId, serviceName, clusterName *string) (*consoleclient.ServiceDeploymentExtended, error)
	DeleteClusterService(serviceId string) (*consoleclient.DeleteServiceDeployment, error)
}

func NewConsoleClient(token, url string) (ConsoleClient, error) {
	return &consoleClient{
		client: consoleclient.NewClient(http.DefaultClient, fmt.Sprintf("%s/gql", url), func(req *http.Request) {
			req.Header.Set("Authorization", fmt.Sprintf("Token %s", token))
		}),
		ctx: context.Background(),
	}, nil
}
