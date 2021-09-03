package bundle

import (
	"strings"
	"fmt"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/utils"
)

func configureOidc(repo string, client *api.Client, recipe *api.Recipe, ctx map[string]interface{}) error {
	if recipe.OidcSettings == nil {
		return nil
	}

	confirm, err := utils.ReadLine("Do you want to enable plural OIDC? (yN)")
	if confirm != "y" || err != nil {
		return err
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

	return client.OIDCProvider(inst.Id, oidcSettings)
}

func formatRedirectUri(settings *api.OIDCSettings, ctx map[string]interface{}) (string, error) {
	domain, ok := ctx[settings.DomainKey]
	if !ok {
		return "", fmt.Errorf("No domain setting for %s in context", settings.DomainKey)
	}

	return strings.ReplaceAll(settings.UriFormat, "{domain}", domain.(string)), nil
}