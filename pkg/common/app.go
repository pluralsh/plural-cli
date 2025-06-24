package common

import (
	"fmt"

	"github.com/pluralsh/plural-cli/pkg/up"

	"github.com/urfave/cli"
)

func HandleDown(c *cli.Context) error {
	if !Affirm(AffirmDown, "PLURAL_DOWN_AFFIRM_DESTROY") {
		return fmt.Errorf("cancelled destroy")
	}

	ctx, err := up.Build(c.Bool("cloud"))
	if err != nil {
		return err
	}

	return ctx.Destroy()
}
