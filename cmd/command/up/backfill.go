package up

import (
	"fmt"
	"strings"

	"encoding/base64"

	"github.com/pluralsh/gqlclient"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/crypto"
	"github.com/samber/lo"
)

func (p *Plural) backfillEncryption() error {
	instances, err := p.Plural.Client.GetConsoleInstances()
	if err != nil {
		return err
	}

	conf := console.ReadConfig()

	if conf.Url == "" {
		return fmt.Errorf("You haven't configured your Plural Console client yet")
	}

	var id string
	for _, inst := range instances {
		if strings.Contains(conf.Url, inst.URL) {
			id = inst.ID
		}
	}
	if id == "" {
		return fmt.Errorf("Your configuration doesn't match to any existing Plural Console")
	}

	prov, err := crypto.Build()
	if err != nil {
		return err
	}

	raw, err := prov.SymmetricKey()
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(raw)

	return p.Plural.Client.UpdateConsoleInstance(id, gqlclient.ConsoleInstanceUpdateAttributes{
		Configuration: &gqlclient.ConsoleConfigurationUpdateAttributes{
			EncryptionKey: lo.ToPtr(encoded),
		},
	})
}
