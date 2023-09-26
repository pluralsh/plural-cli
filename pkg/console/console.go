package console

import (
	"context"
	"fmt"
	"net/http"

	consoleclient "github.com/pluralsh/console-client-go"
)

type consoleClient struct {
	ctx          context.Context
	pluralClient *consoleclient.Client
}

type ConsoleClient interface {
	ListClusters() ([]Cluster, error)
	ListClusterServices(clusterId string) ([]ServiceDeployment, error)
	CreateRepository(url string, privateKey, passphrase, username, password *string) (*GitRepository, error)
	ListRepositories() ([]GitRepository, error)
}

func NewConsoleClient(token, url string) (ConsoleClient, error) {
	return &consoleClient{
		pluralClient: consoleclient.NewClient(http.DefaultClient, fmt.Sprintf("%s/gql", url), func(req *http.Request) {
			req.Header.Set("Authorization", fmt.Sprintf("Token %s", token))
		}),
		ctx: context.Background(),
	}, nil
}
