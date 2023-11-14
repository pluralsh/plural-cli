package plural

import (
	"io"
	"os"

	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/urfave/cli"
)

func configCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "amend",
			Usage:     "modify config",
			ArgsUsage: "[key] [value]",
			Action:    latestVersion(handleAmend),
		},
		{
			Name:      "read",
			Usage:     "dumps config",
			ArgsUsage: "",
			Action:    latestVersion(handleRead),
		},
		{
			Name:   "import",
			Usage:  "imports a new config from a given token",
			Action: latestVersion(handleConfigImport),
		},
	}
}

func profileCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "use",
			Usage:     "moves the config in PROFILE to the current config",
			ArgsUsage: "PROFILE",
			Action:    latestVersion(handleUseProfile),
		},
		{
			Name:      "save",
			Usage:     "saves the current config as PROFILE",
			ArgsUsage: "PROFILE",
			Action:    latestVersion(handleSaveProfile),
		},
		{
			Name:      "show",
			Usage:     "displays the configuration for the current profile",
			ArgsUsage: "PROFILE",
			Action:    latestVersion(handleRead),
		},
		{
			Name:      "list",
			Usage:     "lists all saved profiles",
			ArgsUsage: "",
			Action:    latestVersion(listProfiles),
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

func handleUseProfile(c *cli.Context) error {
	return config.Profile(c.Args().Get(0))
}

func handleSaveProfile(c *cli.Context) error {
	conf := config.Read()
	return conf.SaveProfile(c.Args().Get(0))
}

func listProfiles(c *cli.Context) error {
	profiles, err := config.Profiles()
	if err != nil {
		return err
	}

	headers := []string{"Name", "Email", "Endpoint"}
	return utils.PrintTable(profiles, headers, func(profile *config.VersionedConfig) ([]string, error) {
		return []string{profile.Metadata.Name, profile.Spec.Email, profile.Spec.BaseUrl()}, nil
	})
}
