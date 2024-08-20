package down

import (
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/urfave/cli"
)

func Command() cli.Command {
	return cli.Command{
		Name:   "down",
		Usage:  "destroys your management cluster and any apps installed on it",
		Action: common.LatestVersion(common.HandleDown),
	}
}
