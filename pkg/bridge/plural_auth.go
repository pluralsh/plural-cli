package bridge

import (
	"context"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/config"
)

// PluralAuthFactory adapts the legacy GraphQL client to AuthClientFactory.
type PluralAuthFactory struct{}

func (PluralAuthFactory) New(ctx context.Context, endpoint, credential string) AuthClient {
	conf := &config.Config{Endpoint: endpoint, Token: credential}
	return pluralAuthClient{client: api.FromConfigWithContext(ctx, conf)}
}

type pluralAuthClient struct{ client api.Client }

func (c pluralAuthClient) DeviceLogin(context.Context) (DeviceAuthorization, error) {
	device, err := c.client.DeviceLogin()
	if err != nil {
		return DeviceAuthorization{}, err
	}
	return DeviceAuthorization{LoginURL: device.LoginUrl, DeviceToken: device.DeviceToken}, nil
}

func (c pluralAuthClient) PollLoginToken(_ context.Context, deviceToken string) (string, error) {
	return c.client.PollLoginToken(deviceToken)
}

func (c pluralAuthClient) CurrentIdentity(context.Context) (string, error) {
	me, err := c.client.Me()
	if err != nil {
		return "", err
	}
	return me.Email, nil
}

func (c pluralAuthClient) ImpersonateServiceAccount(_ context.Context, email string) (string, string, error) {
	return c.client.ImpersonateServiceAccount(email)
}

func (c pluralAuthClient) GrabAccessToken(context.Context) (string, error) {
	return c.client.GrabAccessToken()
}
