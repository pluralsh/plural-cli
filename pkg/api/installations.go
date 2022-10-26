package api

import (
	"fmt"

	"github.com/pluralsh/gqlclient"
	"github.com/pluralsh/gqlclient/pkg/utils"
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

func (client *client) GetInstallation(name string) (*Installation, error) {
	resp, err := client.pluralClient.GetInstallation(client.ctx, &name)
	if err != nil {
		return nil, err
	}

	return convertInstallation(resp.Installation), nil

}

func (client *client) GetInstallationById(id string) (*Installation, error) {
	resp, err := client.pluralClient.GetInstallationByID(client.ctx, &id)
	if err != nil {
		return nil, err
	}
	return convertInstallation(resp.Installation), nil
}

func convertInstallation(installation *gqlclient.InstallationFragment) *Installation {
	if installation == nil {
		return nil
	}
	i := &Installation{
		Id:         installation.ID,
		LicenseKey: utils.ConvertStringPointer(installation.LicenseKey),
		Context:    installation.Context,
		AcmeKeyId:  utils.ConvertStringPointer(installation.AcmeKeyID),
		AcmeSecret: utils.ConvertStringPointer(installation.AcmeSecret),
	}
	if installation.Repository != nil {
		i.Repository = &Repository{
			Id:          installation.Repository.ID,
			Name:        installation.Repository.Name,
			Description: utils.ConvertStringPointer(installation.Repository.Description),
			Icon:        utils.ConvertStringPointer(installation.Repository.Icon),
			DarkIcon:    utils.ConvertStringPointer(installation.Repository.DarkIcon),
			Notes:       utils.ConvertStringPointer(installation.Repository.Notes),
			Recipes:     []*Recipe{},
		}
		if installation.Repository.Publisher != nil {
			i.Repository.Publisher = &Publisher{
				Name: installation.Repository.Publisher.Name,
			}
		}
		for _, recipe := range installation.Repository.Recipes {
			i.Repository.Recipes = append(i.Repository.Recipes, &Recipe{
				Name: recipe.Name,
			})
		}
	}

	if installation.OidcProvider != nil {
		i.OIDCProvider = &OIDCProvider{
			Id:           installation.OidcProvider.ID,
			ClientId:     installation.OidcProvider.ClientID,
			ClientSecret: installation.OidcProvider.ClientSecret,
			RedirectUris: utils.ConvertStringArrayPointer(installation.OidcProvider.RedirectUris),
			Bindings:     []*ProviderBinding{},
		}
		if installation.OidcProvider.Configuration != nil {
			i.OIDCProvider.Configuration = &OAuthConfiguration{
				Issuer:                utils.ConvertStringPointer(installation.OidcProvider.Configuration.Issuer),
				AuthorizationEndpoint: utils.ConvertStringPointer(installation.OidcProvider.Configuration.AuthorizationEndpoint),
				TokenEndpoint:         utils.ConvertStringPointer(installation.OidcProvider.Configuration.TokenEndpoint),
				JwksUri:               utils.ConvertStringPointer(installation.OidcProvider.Configuration.JwksURI),
				UserinfoEndpoint:      utils.ConvertStringPointer(installation.OidcProvider.Configuration.UserinfoEndpoint),
			}
		}
		for _, binding := range installation.OidcProvider.Bindings {
			pb := &ProviderBinding{}
			if binding.User != nil {
				pb.User = &User{
					Id:    binding.User.ID,
					Email: binding.User.Email,
				}
			}
			if binding.Group != nil {
				pb.Group = &Group{
					Id:   binding.Group.ID,
					Name: binding.Group.Name,
				}
			}
			i.OIDCProvider.Bindings = append(i.OIDCProvider.Bindings, pb)
		}
	}

	return i
}

func (client *client) GetInstallations() ([]*Installation, error) {
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

func (client *client) OIDCProvider(id string, attributes *OidcProviderAttributes) error {
	bindings := make([]*gqlclient.BindingAttributes, 0)
	for _, bind := range attributes.Bindings {
		groupId := bind.GroupId
		userId := bind.UserId
		bindings = append(bindings, &gqlclient.BindingAttributes{
			GroupID: &groupId,
			UserID:  &userId,
		})
	}

	redirectUris := convertRedirectUris(attributes.RedirectUris)
	_, err := client.pluralClient.UpsertOidcProvider(client.ctx, id, gqlclient.OidcAttributes{
		AuthMethod:   gqlclient.OidcAuthMethod(attributes.AuthMethod),
		Bindings:     bindings,
		RedirectUris: redirectUris,
	})
	fmt.Printf("%+v", redirectUris)
	return err
}

func (client *client) ResetInstallations() (int, error) {
	resp, err := client.pluralClient.ResetInstallations(client.ctx)
	if err != nil {
		return 0, err
	}

	return int(*resp.ResetInstallations), err
}

func convertRedirectUris(uris []string) []*string {
	res := make([]*string, len(uris))
	for i := range uris {
		res[i] = &uris[i]
	}
	return res
}
