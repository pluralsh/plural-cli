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
	ListClusterServices(clusterId string) (*consoleclient.ListServiceDeployment, error)
	CreateRepository(url string, privateKey, passphrase, username, password *string) (*consoleclient.CreateGitRepository, error)
	ListRepositories() (*consoleclient.ListGitRepositories, error)
	CreateClusterService(clusterId string, attr consoleclient.ServiceDeploymentAttributes) (*consoleclient.CreateServiceDeployment, error)
	UpdateClusterService(serviceId string, attr consoleclient.ServiceUpdateAttributes) (*consoleclient.UpdateServiceDeployment, error)
	GetClusterService(serviceId string) (*consoleclient.GetServiceDeployment, error)
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
