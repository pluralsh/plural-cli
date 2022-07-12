package api

import (
	"github.com/pluralsh/gqlclient"
)

type Binding struct {
	UserId  string
	GroupId string
}

type OidcProviderAttributes struct {
	RedirectUris []string
	AuthMethod   string
	Bindings     []Binding
}

func (client *Client) GetInstallation(name string) (*Installation, error) {
	resp, err := client.pluralClient.GetInstallation(client.ctx, &name)
	if err != nil {
		return nil, err
	}

	return convertInstallation(resp.Installation), nil

}

func (client *Client) GetInstallationById(id string) (*Installation, error) {
	resp, err := client.pluralClient.GetInstallationByID(client.ctx, &id)
	if err != nil {
		return nil, err
	}
	return convertInstallation(resp.Installation), nil
}

func convertInstallation(installation *gqlclient.InstallationFragment) *Installation {
	i := &Installation{
		Id: installation.ID,
		Repository: &Repository{
			Id:   installation.Repository.ID,
			Name: installation.Repository.Name,
			Publisher: &Publisher{
				Name: installation.Repository.Publisher.Name,
			},
		},
	}
	if installation.Repository.DarkIcon != nil {
		i.Repository.DarkIcon = *installation.Repository.DarkIcon
	}

	if installation.Repository.Description != nil {
		i.Repository.Description = *installation.Repository.Description
	}

	if installation.Repository.Icon != nil {
		i.Repository.Icon = *installation.Repository.Icon
	}

	if installation.Repository.Notes != nil {
		i.Repository.Notes = *installation.Repository.Notes
	}

	if installation.LicenseKey != nil {
		i.LicenseKey = *installation.LicenseKey
	}
	if installation.AcmeKeyID != nil {
		i.AcmeKeyId = *installation.AcmeKeyID
	}
	if installation.AcmeSecret != nil {
		i.AcmeSecret = *installation.AcmeSecret
	}

	return i
}

func (client *Client) GetInstallations() ([]*Installation, error) {
	result := make([]*Installation, 0)

	resp, err := client.pluralClient.GetInstallations(client.ctx)
	if err != nil {
		return result, err
	}

	for _, edge := range resp.Installations.Edges {
		result = append(result, convertInstallation(edge.Node))
	}

	return result, err
}

func (client *Client) OIDCProvider(id string, attributes *OidcProviderAttributes) error {
	var groupId = attributes.Bindings[0].GroupId
	var userId = attributes.Bindings[0].UserId
	_, err := client.pluralClient.UpsertOidcProvider(client.ctx, id, gqlclient.OidcAttributes{
		AuthMethod: gqlclient.OidcAuthMethod(attributes.AuthMethod),
		Bindings: []*gqlclient.BindingAttributes{
			{
				GroupID: &groupId,
				UserID:  &userId,
			},
		},
		RedirectUris: convertRedirectUris(attributes.RedirectUris),
	})
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) ResetInstallations() (int, error) {
	resp, err := client.pluralClient.ResetInstallations(client.ctx)
	if err != nil {
		return 0, err
	}

	return int(*resp.ResetInstallations), err
}

func convertRedirectUris(uri []string) []*string {
	res := make([]*string, 0)
	for _, s := range uri {
		res = append(res, &s)
	}
	return res
}
