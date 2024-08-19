package main

import (
	"github.com/fatih/color"
	"github.com/pluralsh/plural-cli/cmd/command/cd"
	cryptocmd "github.com/pluralsh/plural-cli/cmd/command/crypto"
	cmdinit "github.com/pluralsh/plural-cli/cmd/command/init"
	"github.com/pluralsh/plural-cli/cmd/command/mgmt"
	"github.com/pluralsh/plural-cli/cmd/command/pr"
	"github.com/pluralsh/plural-cli/cmd/command/stack"
	"github.com/pluralsh/plural-cli/cmd/command/up"
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	conf "github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/crypto"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/urfave/cli"
	"helm.sh/helm/v3/pkg/action"
	"log"
	"os"
)

const ApplicationName = "pluralctl"

type Plural struct {
	client.Plural
	HelmConfiguration *action.Configuration
}

func (p *Plural) getCommands() []cli.Command {
	return []cli.Command{
		{
			Name:    "version",
			Aliases: []string{"v", "vsn"},
			Usage:   "Gets cli version info",
			Action:  common.VersionInfo,
		},
		{
			Name:   "down",
			Usage:  "destroys your management cluster and any apps installed on it",
			Action: common.LatestVersion(common.HandleDown),
		},
		{
			Name:   "login",
			Usage:  "logs into plural and saves credentials to the current config profile",
			Action: common.LatestVersion(common.HandleLogin),
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "endpoint",
					Usage: "the endpoint for the plural installation you're working with",
				},
				cli.StringFlag{
					Name:  "service-account",
					Usage: "email for the service account you'd like to use for this workspace",
				},
			},
			Category: "User Profile",
		},
	}
}

func globalFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "profile-file",
			Usage:       "configure your config.yml profile `FILE`",
			EnvVar:      "PLURAL_PROFILE_FILE",
			Destination: &conf.ProfileFile,
		},
		cli.StringFlag{
			Name:        "encryption-key-file",
			Usage:       "configure your encryption key `FILE`",
			EnvVar:      "PLURAL_ENCRYPTION_KEY_FILE",
			Destination: &crypto.EncryptionKeyFile,
		},
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "enable debug mode",
			EnvVar:      "PLURAL_DEBUG_ENABLE",
			Destination: &utils.EnableDebug,
		},
	}
}

func main() {

	plural := &Plural{}

	app := cli.NewApp()
	app.Name = ApplicationName
	app.Usage = "Tooling to manage and operate a fleet of clusters"
	app.EnableBashCompletion = true
	app.Flags = globalFlags()
	commands := []cli.Command{
		cryptocmd.Command(plural.Plural),
		cd.Command(plural.Plural, plural.HelmConfiguration),
		up.Command(plural.Plural),
		pr.Command(),
		stack.Command(plural.Plural),
		cmdinit.Command(plural.Plural),
		mgmt.Command(plural.Plural),
	}
	commands = append(commands, plural.getCommands()...)
	app.Commands = commands
	if os.Getenv("ENABLE_COLOR") != "" {
		color.NoColor = false
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
