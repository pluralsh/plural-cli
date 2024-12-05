package common

import (
	"fmt"

	"github.com/pluralsh/plural-cli/pkg/up"

	"github.com/urfave/cli"
)

func HandleDown(_ *cli.Context) error {
	if !Affirm(AffirmDown, "PLURAL_DOWN_AFFIRM_DESTROY") {
		return fmt.Errorf("cancelled destroy")
	}

	ctx, err := up.Build(false)
	if err != nil {
		return err
	}

	return ctx.Destroy()
}
