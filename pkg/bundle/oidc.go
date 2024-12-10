package bundle

import (
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

func SetupOIDC(repo string, client api.Client, redirectUris []string, authMethod string) error {
	inst, err := client.GetInstallation(repo)
	if err != nil {
		return api.GetErrorResponse(err, "GetInstallation")
	}

	me, err := client.Me()
	if err != nil {
		return api.GetErrorResponse(err, "Me")
	}

	oidcSettings := &api.OidcProviderAttributes{
		RedirectUris: redirectUris,
		AuthMethod:   authMethod,
		Bindings: []api.Binding{
			{UserId: me.Id},
		},
	}
	mergeOidcAttributes(inst, oidcSettings)
	err = client.OIDCProvider(inst.Id, oidcSettings)
	return api.GetErrorResponse(err, "OIDCProvider")
}

func mergeOidcAttributes(inst *api.Installation, attributes *api.OidcProviderAttributes) {
	if inst.OIDCProvider == nil {
		return
	}

	provider := inst.OIDCProvider
	attributes.RedirectUris = utils.Dedupe(append(attributes.RedirectUris, provider.RedirectUris...))
	bindings := attributes.Bindings
	for _, val := range provider.Bindings {
		// attributes is only pre-populated with the current user right now
		if val.User != nil && val.User.Id != attributes.Bindings[0].UserId {
			bindings = append(bindings, api.Binding{UserId: val.User.Id})
		} else if val.Group != nil {
			bindings = append(bindings, api.Binding{GroupId: val.Group.Id})
		}
	}
	attributes.Bindings = bindings
}
