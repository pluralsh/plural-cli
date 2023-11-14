package scaffold

import (
	"bytes"
	"fmt"

	"github.com/pluralsh/plural-cli/pkg/utils"
)

const filterTmpl = "%s filter=plural-crypt diff=plural-crypt\n"

func buildSecrets(file string, secrets []string) error {
	var b bytes.Buffer
	b.Grow(32)
	for _, secret := range secrets {
		fmt.Fprintf(&b, filterTmpl, secret)
	}

	return utils.WriteFile(file, b.Bytes())
}
