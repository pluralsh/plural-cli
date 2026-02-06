package up

import (
	"fmt"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/console"
)

func (p *Plural) ValidateConsoleConfig() error {
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

	return nil
}
