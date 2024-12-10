package version

import (
	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/urfave/cli"
)

func Command() cli.Command {
	return cli.Command{
		Name:    "version",
		Aliases: []string{"v", "vsn"},
		Usage:   "gets plural cli version info",
		Action:  common.VersionInfo,
	}
}
