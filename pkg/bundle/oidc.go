package bundle

import (
	"strings"
	"fmt"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/AlecAivazis/survey/v2"
)

var oidcConfirmed bool

func configureOidc(repo string, client *api.Client, recipe *api.Recipe, ctx map[string]interface{}, confirm *bool) error {
	if recipe.OidcSettings == nil {
		return nil
	}

	confirmOidc(confirm)

	if !*confirm {
		return nil
	}

	settings := recipe.OidcSettings
	redirectUri, err := formatRedirectUri(settings, ctx)
	if err != nil {
		return err
	}

	inst, err := client.GetInstallation(repo)
	if err != nil {
		return err
	}

	me, err := client.Me()
	if err != nil {
		return err
	}

	oidcSettings := &api.OidcProviderAttributes{
		RedirectUris: []string{ redirectUri },
		AuthMethod: settings.AuthMethod,
		Bindings: []api.Binding{
			{ UserId: me.Id },
		},
	}
	mergeOidcAttributes(inst, oidcSettings)

	return client.OIDCProvider(inst.Id, oidcSettings)
}


func mergeOidcAttributes(inst *api.Installation, attributes *api.OidcProviderAttributes) {
	if (inst.OIDCProvider == nil) {
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

func formatRedirectUri(settings *api.OIDCSettings, ctx map[string]interface{}) (string, error) {
	domain, ok := ctx[settings.DomainKey]
	if !ok {
		return "", fmt.Errorf("No domain setting for %s in context", settings.DomainKey)
	}

	return strings.ReplaceAll(settings.UriFormat, "{domain}", domain.(string)), nil
}

func confirmOidc(confirm *bool) {
	if oidcConfirmed {
		return
	}

	survey.AskOne(&survey.Confirm{
		Message: "Enable plural OIDC",
		Default: true,
	}, confirm, survey.WithValidator(survey.Required))

	oidcConfirmed = true
}