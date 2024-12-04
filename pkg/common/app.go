package common

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/up"

	"github.com/urfave/cli"

	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

func HandleInfo(c *cli.Context) error {
	repo := c.Args().Get(0)
	conf := config.Read()

	_, err := exec.LookPath("k9s")
	if err != nil {
		utils.LogError().Println(err)
		if strings.Contains(err.Error(), exec.ErrNotFound.Error()) {
			utils.Error("Application k9s not installed.\n")
			fmt.Println("Please install it first from here: https://k9scli.io/topics/install/ and try again")
			return nil
		}
	}

	cmd := exec.Command("k9s", "-n", conf.Namespace(repo))
	return cmd.Run()
}

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
