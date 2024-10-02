package profile

import (
	"os"

	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/urfave/cli"
)

func Command() cli.Command {
	return cli.Command{
		Name:        "profile",
		Usage:       "Commands for managing config profiles for plural",
		Subcommands: profileCommands(),
		Category:    "User Profile",
	}
}

func profileCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "use",
			Usage:     "moves the config in PROFILE to the current config",
			ArgsUsage: "{profile}",
			Action:    common.LatestVersion(common.RequireArgs(handleUseProfile, []string{"{profile}"})),
		},
		{
			Name:      "save",
			Usage:     "saves the current config as PROFILE",
			ArgsUsage: "{profile}",
			Action:    common.LatestVersion(common.RequireArgs(handleSaveProfile, []string{"{profile}"})),
		},
		{
			Name:   "show",
			Usage:  "displays the configuration for the current profile",
			Action: common.LatestVersion(handleRead),
		},
		{
			Name:      "list",
			Usage:     "lists all saved profiles",
			ArgsUsage: "",
			Action:    common.LatestVersion(listProfiles),
		},
	}
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

func handleRead(c *cli.Context) error {
	conf := config.Read()
	d, err := conf.Marshal()
	if err != nil {
		return err
	}

	os.Stdout.Write(d)
	return nil
}
