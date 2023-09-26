package console

import (
	"context"
	"fmt"
	consoleclient "github.com/pluralsh/console-client-go"
	"net/http"
)

type consoleClient struct {
	ctx          context.Context
	pluralClient *consoleclient.Client
}

type ConsoleClient interface {
	ListClusters() ([]Cluster, error)
}

func NewConsoleClient(token, url string) (ConsoleClient, error) {
	authHeader := func(req *http.Request) {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}
	return &consoleClient{
		pluralClient: consoleclient.NewClient(http.DefaultClient, fmt.Sprintf("%s/gql", url), authHeader),
		ctx:          context.Background(),
	}, nil
}
