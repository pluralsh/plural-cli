package api

import (
	"context"
	"fmt"

	"github.com/michaeljguarino/graphql"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/utils"
)

const (
	pageSize = 100
)

type Client struct {
	gqlClient *graphql.Client
	config    config.Config
}

func NewClient() *Client {
	conf := config.Read()
	return FromConfig(&conf)
}

func FromConfig(conf *config.Config) *Client {
	return &Client{graphql.NewClient(conf.Url()), *conf}
}

func NewUploadClient() *Client {
	conf := config.Read()
	client := graphql.NewClient(conf.Url(), graphql.UseMultipartForm())
	return &Client{client, conf}
}

func (client *Client) Build(doc string) *graphql.Request {
	req := graphql.NewRequest(doc)
	req.Header.Set("Authorization", "Bearer "+client.config.Token)
	return req
}

func (client *Client) Run(req *graphql.Request, resp interface{}) error {
	err := client.gqlClient.Run(context.Background(), req, &resp)
	return utils.HighlightError(err)
}

func (client *Client) EnableLogging() {
	client.gqlClient.Log = func(l string) { fmt.Println(l) }
}
