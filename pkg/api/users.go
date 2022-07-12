package api

import (
	"fmt"
	"strings"

	"github.com/pluralsh/gqlclient"
)

type UpgradeAttributes struct {
	Message string
}

type UserEventAttributes struct {
	Event  string
	Data   string
	Status string
}

type DeviceLogin struct {
	LoginUrl    string
	DeviceToken string
}

type LoginMethod struct {
	LoginMethod string
	Token       string
}

type Me struct {
	Id    string
	Email string
}

func (client *Client) Me() (*Me, error) {
	resp, err := client.pluralClient.Me(client.ctx)
	if err != nil {
		return nil, err
	}
	return &Me{
		Id:    resp.Me.ID,
		Email: resp.Me.Email,
	}, nil
}

func (client *Client) LoginMethod(email string) (*LoginMethod, error) {
	resp, err := client.pluralClient.GetLoginMethod(client.ctx, email)
	if err != nil {
		return nil, err
	}
	return &LoginMethod{
		LoginMethod: string(resp.LoginMethod.LoginMethod),
		Token:       *resp.LoginMethod.Token,
	}, nil
}

func (client *Client) PollLoginToken(token string) (string, error) {
	resp, err := client.pluralClient.PollLoginToken(client.ctx, token)
	if err != nil {
		return "", err
	}

	if resp.LoginToken != nil && resp.LoginToken.Jwt != nil {
		return *resp.LoginToken.Jwt, err
	}

	return "", fmt.Errorf("the JWT token is empty")
}

func (client *Client) DeviceLogin() (*DeviceLogin, error) {
	resp, err := client.pluralClient.DevLogin(client.ctx)
	if err != nil {
		return nil, err
	}

	if resp.DeviceLogin != nil {
		return &DeviceLogin{
			LoginUrl:    resp.DeviceLogin.LoginURL,
			DeviceToken: resp.DeviceLogin.DeviceToken,
		}, nil
	}

	return nil, fmt.Errorf("the response DeviceLogin is nil")
}

func (client *Client) Login(email, pwd string) (string, error) {
	resp, err := client.pluralClient.Login(client.ctx, email, pwd)
	if err != nil {
		return "", err
	}

	if resp.Login != nil && resp.Login.Jwt != nil {
		return *resp.Login.Jwt, nil
	}
	return "", fmt.Errorf("the JWT token is empty")
}

func (client *Client) ImpersonateServiceAccount(email string) (string, string, error) {
	resp, err := client.pluralClient.ImpersonateServiceAccount(client.ctx, &email)
	if err != nil {
		return "", "", err
	}
	if resp.ImpersonateServiceAccount != nil && resp.ImpersonateServiceAccount.Jwt != nil {
		return *resp.ImpersonateServiceAccount.Jwt, resp.ImpersonateServiceAccount.Email, err
	}
	return "", "", fmt.Errorf("the response ImpersonateServiceAccount is nil")
}

func (client *Client) CreateAccessToken() (string, error) {
	resp, err := client.pluralClient.CreateAccessToken(client.ctx)
	if err != nil {
		return "", err
	}
	return *resp.CreateToken.Token, err
}

func (client *Client) GrabAccessToken() (string, error) {
	resp, err := client.pluralClient.ListTokens(client.ctx)
	if err != nil {
		return "", err
	}

	if len(resp.Tokens.Edges) > 0 {
		return *resp.Tokens.Edges[0].Node.Token, nil
	}

	return client.CreateAccessToken()
}

func (client *Client) ListKeys(emails []string) ([]*PublicKey, error) {
	emailsInput := make([]*string, 0)
	for _, email := range emails {
		emailsInput = append(emailsInput, &email)
	}

	resp, err := client.pluralClient.ListKeys(client.ctx, emailsInput)
	if err != nil {
		return nil, err
	}

	keys := make([]*PublicKey, 0)
	for _, edge := range resp.PublicKeys.Edges {
		keys = append(keys, &PublicKey{
			Id:      edge.Node.ID,
			Content: edge.Node.Content,
			User: &User{
				Id:    edge.Node.User.ID,
				Email: edge.Node.User.Email,
				Name:  edge.Node.User.Name,
			},
		})
	}
	return keys, nil
}

func (client *Client) CreateKey(name, content string) error {
	_, err := client.pluralClient.CreateKey(client.ctx, content, name)
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) GetEabCredential(cluster, provider string) (*EabCredential, error) {
	resp, err := client.pluralClient.GetEabCredential(client.ctx, cluster, gqlclient.Provider(toProvider(provider)))
	if err != nil {
		return nil, err
	}

	return &EabCredential{
		Id:       resp.EabCredential.ID,
		KeyId:    resp.EabCredential.KeyID,
		HmacKey:  resp.EabCredential.HmacKey,
		Cluster:  resp.EabCredential.Cluster,
		Provider: string(resp.EabCredential.Provider),
	}, nil
}

func (client *Client) DeleteEabCredential(cluster, provider string) error {
	_, err := client.pluralClient.DeleteEabCredential(client.ctx, cluster, gqlclient.Provider(toProvider(provider)))
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) CreateEvent(event *UserEventAttributes) error {
	status := gqlclient.UserEventStatus(event.Status)
	_, err := client.pluralClient.CreateEvent(client.ctx, gqlclient.UserEventAttributes{
		Data:   &event.Data,
		Event:  event.Event,
		Status: &status,
	})
	if err != nil {
		return err
	}

	return nil
}

func toProvider(prov string) string {
	if prov == "google" {
		return "GCP"
	}

	return strings.ToUpper(prov)
}
