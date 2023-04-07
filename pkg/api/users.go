package api

import (
	"fmt"

	"github.com/pluralsh/gqlclient"
	"github.com/pluralsh/polly/algorithms"
	"github.com/samber/lo"
)

type UserEventAttributes struct {
	Event  string
	Data   string
	Status string
}

type KeyBackupAttributes struct {
	Name         string
	Repositories []string
	Key          string
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
	Id      string
	Email   string
	Demoing bool
}

func (client *client) Me() (*Me, error) {
	resp, err := client.pluralClient.Me(client.ctx)
	if err != nil {
		return nil, err
	}
	return &Me{
		Id:      resp.Me.ID,
		Email:   resp.Me.Email,
		Demoing: lo.FromPtr(resp.Me.Demoing),
	}, nil
}

func (client *client) LoginMethod(email string) (*LoginMethod, error) {
	resp, err := client.pluralClient.GetLoginMethod(client.ctx, email)
	if err != nil {
		return nil, err
	}
	return &LoginMethod{
		LoginMethod: string(resp.LoginMethod.LoginMethod),
		Token:       *resp.LoginMethod.Token,
	}, nil
}

func (client *client) PollLoginToken(token string) (string, error) {
	resp, err := client.pluralClient.PollLoginToken(client.ctx, token)
	if err != nil {
		return "", err
	}

	if resp.LoginToken != nil && resp.LoginToken.Jwt != nil {
		return *resp.LoginToken.Jwt, err
	}

	return "", fmt.Errorf("the JWT token is empty")
}

func (client *client) DeviceLogin() (*DeviceLogin, error) {
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

func (client *client) Login(email, pwd string) (string, error) {
	resp, err := client.pluralClient.Login(client.ctx, email, pwd)
	if err != nil {
		return "", err
	}

	if resp.Login != nil && resp.Login.Jwt != nil {
		return *resp.Login.Jwt, nil
	}
	return "", fmt.Errorf("the JWT token is empty")
}

func (client *client) ImpersonateServiceAccount(email string) (string, string, error) {
	resp, err := client.pluralClient.ImpersonateServiceAccount(client.ctx, &email)
	if err != nil {
		return "", "", err
	}
	if resp.ImpersonateServiceAccount != nil && resp.ImpersonateServiceAccount.Jwt != nil {
		return *resp.ImpersonateServiceAccount.Jwt, resp.ImpersonateServiceAccount.Email, err
	}
	return "", "", fmt.Errorf("the response ImpersonateServiceAccount is nil")
}

func (client *client) CreateAccessToken() (string, error) {
	resp, err := client.pluralClient.CreateAccessToken(client.ctx)
	if err != nil {
		return "", err
	}
	return *resp.CreateToken.Token, err
}

func (client *client) GrabAccessToken() (string, error) {
	resp, err := client.pluralClient.ListTokens(client.ctx)
	if err != nil {
		return "", err
	}

	if len(resp.Tokens.Edges) > 0 {
		return *resp.Tokens.Edges[0].Node.Token, nil
	}

	return client.CreateAccessToken()
}

func (client *client) ListKeys(emails []string) ([]*PublicKey, error) {
	emailsInput := lo.ToSlicePtr(emails)

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

func (client *client) CreateKey(name, content string) error {
	_, err := client.pluralClient.CreateKey(client.ctx, content, name)
	if err != nil {
		return err
	}
	return nil
}

func (client *client) GetEabCredential(cluster, provider string) (*EabCredential, error) {
	resp, err := client.pluralClient.GetEabCredential(client.ctx, cluster, gqlclient.Provider(NormalizeProvider(provider)))
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

func (client *client) DeleteEabCredential(cluster, provider string) error {
	_, err := client.pluralClient.DeleteEabCredential(client.ctx, cluster, gqlclient.Provider(NormalizeProvider(provider)))
	if err != nil {
		return err
	}
	return nil
}

func (client *client) CreateEvent(event *UserEventAttributes) error {
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

func (client *client) GetHelp(prompt string) (string, error) {
	res, err := client.pluralClient.GetHelp(client.ctx, prompt)
	if err != nil {
		return "", err
	}

	return lo.FromPtr(res.HelpQuestion), nil
}

func (client *client) CreateKeyBackup(attrs KeyBackupAttributes) error {
	converted := gqlclient.KeyBackupAttributes{
		Name:         attrs.Name,
		Key:          attrs.Key,
		Repositories: lo.ToSlicePtr(attrs.Repositories),
	}
	_, err := client.pluralClient.CreateBackup(client.ctx, converted)
	return err
}

func (client *client) ListKeyBackups() ([]*KeyBackup, error) {
	resp, err := client.pluralClient.Backups(client.ctx, nil)
	if err != nil {
		return nil, err
	}

	backups := algorithms.Map(resp.KeyBackups.Edges, func(edge *struct {
		Node *gqlclient.KeyBackupFragment "json:\"node\" graphql:\"node\""
	}) *KeyBackup {
		return convertKeyBackup(edge.Node)
	})

	return backups, nil
}

func (client *client) GetKeyBackup(name string) (*KeyBackup, error) {
	resp, err := client.pluralClient.Backup(client.ctx, name)
	if err != nil || resp.KeyBackup == nil {
		return nil, err
	}

	frag := resp.KeyBackup
	return &KeyBackup{
		Name:         frag.Name,
		Digest:       frag.Digest,
		InsertedAt:   lo.FromPtr(frag.InsertedAt),
		Repositories: frag.Repositories,
		Value:        frag.Value,
	}, nil
}

func convertKeyBackup(fragment *gqlclient.KeyBackupFragment) *KeyBackup {
	return &KeyBackup{
		Name:         fragment.Name,
		Digest:       fragment.Digest,
		InsertedAt:   lo.FromPtr(fragment.InsertedAt),
		Repositories: fragment.Repositories,
	}
}
