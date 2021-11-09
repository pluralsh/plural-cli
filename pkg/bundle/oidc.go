package bundle

import (
	"fmt"
	"strings"

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
	redirectUris, err := formatRedirectUris(settings, ctx)
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
		RedirectUris: redirectUris,
		AuthMethod:   settings.AuthMethod,
		Bindings: []api.Binding{
			{UserId: me.Id},
		},
	}

	return client.OIDCProvider(inst.Id, oidcSettings)
}

func formatRedirectUris(settings *api.OIDCSettings, ctx map[string]interface{}) ([]string, error) {
	domains := []string{}
	for index := range settings.DomainKeys {
		domain, ok := ctx[settings.DomainKeys[index]]
		if !ok {
			return []string{""}, fmt.Errorf("No domain setting for %s in context", settings.DomainKeys[index])
		}
		domains = append(domains, strings.ReplaceAll(settings.UriFormat, "{domain}", domain.(string)))
	}

	return domains, nil
}
