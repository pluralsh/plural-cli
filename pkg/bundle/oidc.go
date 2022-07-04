package bundle

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils"
)

var oidcConfirmed bool

func configureOidc(repo string, client *api.Client, recipe *api.Recipe, ctx map[string]interface{}, confirm *bool) error {
	if recipe.OidcSettings == nil {
		return nil
	}

	if err := confirmOidc(confirm); err != nil {
		return err
	}

	if !*confirm {
		return nil
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
	mergeOidcAttributes(inst, oidcSettings)

	return client.OIDCProvider(inst.Id, oidcSettings)
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

func formatRedirectUris(settings *api.OIDCSettings, ctx map[string]interface{}) ([]string, error) {
	res := make([]string, 0)
	domain := ""

	if settings.DomainKey != "" {
		d, ok := ctx[settings.DomainKey]
		if !ok {
			return res, fmt.Errorf("No domain setting for %s in context", settings.DomainKey)
		}

		domain = d.(string)
	}

	proj, err := manifest.FetchProject()
	if err != nil {
		return res, err
	}

	fmtUri := func(uri string) string {
		if domain != "" {
			uri = strings.ReplaceAll(uri, "{domain}", domain)
		}

		if settings.Subdomain {
			uri = strings.ReplaceAll(uri, "{subdomain}", proj.Network.Subdomain)
		}

		return uri
	}

	if settings.UriFormat != "" {
		return []string{fmtUri(settings.UriFormat)}, err
	}

	for _, uri := range settings.UriFormats {
		res = append(res, fmtUri(uri))
	}

	return res, nil
}

func confirmOidc(confirm *bool) error {
	if oidcConfirmed {
		return nil
	}

	if err := survey.AskOne(&survey.Confirm{
		Message: "Enable plural OIDC",
		Default: true,
	}, confirm, survey.WithValidator(survey.Required)); err != nil {
		return err
	}
	oidcConfirmed = true
	return nil
}
