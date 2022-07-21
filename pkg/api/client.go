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

type Client struct {
	ctx          context.Context
	pluralClient *gqlclient.Client
	config       config.Config
}

func NewClient() *Client {
	conf := config.Read()
	return FromConfig(&conf)
}

func FromConfig(conf *config.Config) *Client {
	httpClient := http.Client{
		Transport: &authedTransport{
			key:     conf.Token,
			wrapped: http.DefaultTransport,
		},
	}

	return &Client{
		pluralClient: gqlclient.NewClient(&httpClient, conf.Url()),
		config:       *conf,
		ctx:          context.Background(),
	}
}
