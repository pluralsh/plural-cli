package up

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/pluralsh/gqlclient"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/crypto"
	"github.com/samber/lo"
)

func (p *Plural) backfillEncryption() error {
	instances, err := p.GetConsoleInstances()
	if err != nil {
		return err
	}

	conf := console.ReadConfig()

	if conf.Url == "" {
		return fmt.Errorf("you haven't configured your Plural Console client yet")
	}

	var id string
	for _, inst := range instances {
		if strings.Contains(conf.Url, inst.URL) {
			id = inst.ID
		}
	}
	if id == "" {
		return fmt.Errorf("your configuration doesn't match to any existing Plural Console")
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

	return p.UpdateConsoleInstance(id, gqlclient.ConsoleInstanceUpdateAttributes{
		Configuration: &gqlclient.ConsoleConfigurationUpdateAttributes{
			EncryptionKey: lo.ToPtr(encoded),
		},
	})
}
