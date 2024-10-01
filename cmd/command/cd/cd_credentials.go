package cd

import (
	"fmt"

	"github.com/pluralsh/plural-cli/pkg/common"

	"github.com/pkg/errors"
	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/urfave/cli"
)

func (p *Plural) cdCredentials() cli.Command {
	return cli.Command{
		Name:        "credentials",
		Subcommands: p.cdCredentialsCommands(),
		Usage:       "manage Provider credentials",
	}
}

func (p *Plural) cdCredentialsCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "create",
			ArgsUsage: "{provider-name}",
			Action:    common.LatestVersion(common.RequireArgs(p.handleCreateProviderCredentials, []string{"{provider-name}"})),
			Usage:     "create provider credentials",
		},
		{
			Name:      "delete",
			ArgsUsage: "{id}",
			Action:    common.LatestVersion(common.RequireArgs(p.handleDeleteProviderCredentials, []string{"{id}"})),
			Usage:     "delete provider credentials",
		},
	}
}

func (p *Plural) handleDeleteProviderCredentials(c *cli.Context) error {
	id := c.Args().Get(0)
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	if _, err := p.ConsoleClient.DeleteProviderCredentials(id); err != nil {
		return err
	}
	utils.Success("Provider credential %s has been deleted successfully", id)
	return nil
}

func (p *Plural) handleCreateProviderCredentials(c *cli.Context) error {
	providerName := c.Args().Get(0)
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	attr, err := p.credentialsPreflights()
	if err != nil {
		return err
	}

	resp, err := p.ConsoleClient.CreateProviderCredentials(providerName, *attr)
	if err != nil {
		errList := errors.New("CreateProviderCredentials")
		errList = errors.Wrap(errList, err.Error())
		if *attr.Kind == kindSecret {
			if err := p.SecretDelete(*attr.Namespace, attr.Name); err != nil {
				errList = errors.Wrap(errList, err.Error())
			}
		}
		return errList
	}
	if resp == nil {
		return fmt.Errorf("the response from CreateProviderCredentials is empty")
	}

	headers := []string{"Id", "Name", "Namespace"}
	return utils.PrintTable([]*gqlclient.ProviderCredentialFragment{resp.CreateProviderCredential}, headers, func(sd *gqlclient.ProviderCredentialFragment) ([]string, error) {
		return []string{sd.ID, sd.Name, sd.Namespace}, nil
	})
}
