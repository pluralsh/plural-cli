package plural

import (
	"fmt"
	"strings"

	"github.com/pluralsh/gqlclient"
	"github.com/urfave/cli"

	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/utils"
)

func (p *Plural) authCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "oidc",
			ArgsUsage: "PROVIDER",
			Usage:     "logs in using an exchange from a given oidc id token",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "token",
					Usage: "the oidc id token to use",
				},
				cli.StringFlag{
					Name:  "email",
					Usage: "the plural email you want to log in as",
				},
			},
			Action: p.handleOidcToken,
		},
		{
			Name:        "trust",
			Usage:       "commands to manage oidc trust relationships",
			Subcommands: p.trustCommands(),
		},
	}
}

func (p *Plural) trustCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "list",
			Usage:  "lists all trust relationships attached to the current user",
			Action: p.handleListTrusts,
		},
		{
			Name: "create",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "issuer",
					Usage: "the oidc issuer for this trust relationship",
				},
				cli.StringFlag{
					Name:  "trust",
					Usage: "a regex to establish the claims to trust from this issuer, eg {repo-name}:{ref}:{workflow} for github actions",
				},
			},
			Action: p.handleCreateTrust,
		},
		{
			Name:      "delete",
			ArgsUsage: "ID",
			Usage:     "deletes an existing oidc trust relationship by id",
			Action:    p.handleDeleteTrust,
		},
	}
}

func (p *Plural) handleListTrusts(c *cli.Context) error {
	p.InitPluralClient()
	me, err := p.Client.Me()
	if err != nil {
		return err
	}

	headers := []string{"ID", "Issuer", "Trust", "Created On"}
	return utils.PrintTable(me.TrustRelationships, headers, func(t *gqlclient.OidcTrustRelationshipFragment) ([]string, error) {
		return []string{t.ID, t.Issuer, t.Trust, *t.InsertedAt}, nil
	})
}

func (p *Plural) handleCreateTrust(c *cli.Context) error {
	p.InitPluralClient()
	issuer, trust := c.String("issuer"), c.String("trust")
	return p.Client.CreateTrust(issuer, trust)
}

func (p *Plural) handleDeleteTrust(c *cli.Context) error {
	p.InitPluralClient()
	id := c.Args().Get(0)
	return p.Client.DeleteTrust(id)
}

func (p *Plural) handleOidcToken(c *cli.Context) error {
	p.InitPluralClient()
	prov := c.Args().Get(0)
	token, email := c.String("token"), c.String("email")
	provider := gqlclient.ExternalOidcProvider(strings.ToUpper(prov))
	if !provider.IsValid() {
		return fmt.Errorf("Invalid oidc provider %s", prov)
	}

	token, err := p.Client.OidcToken(provider, token, email)
	if err != nil {
		return err
	}

	conf := config.Config{Token: token, Email: email}
	utils.Success("Logged in as %s", email)
	return conf.Flush()
}
