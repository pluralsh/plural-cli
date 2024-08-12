package config

import (
	"io"
	"os"

	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/urfave/cli"
)

func Command() cli.Command {
	return cli.Command{
		Name:        "config",
		Aliases:     []string{"conf"},
		Usage:       "reads/modifies cli configuration",
		Subcommands: configCommands(),
		Category:    "User Profile",
	}
}

func configCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "amend",
			Usage:     "modify config",
			ArgsUsage: "[key] [value]",
			Action:    common.LatestVersion(handleAmend),
		},
		{
			Name:      "read",
			Usage:     "dumps config",
			ArgsUsage: "",
			Action:    common.LatestVersion(handleRead),
		},
		{
			Name:   "import",
			Usage:  "imports a new config from a given token",
			Action: common.LatestVersion(handleConfigImport),
		},
	}
}

func handleAmend(c *cli.Context) error {
	return config.Amend(c.Args().Get(0), c.Args().Get(1))
}

func handleRead(c *cli.Context) error {
	conf := config.Read()
	d, err := conf.Marshal()
	if err != nil {
		return err
	}

	os.Stdout.Write(d)
	return nil
}

func handleConfigImport(c *cli.Context) error {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	return config.FromToken(string(data))
}
