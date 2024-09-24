package main

import (
	"log"
	"os"

	"github.com/pluralsh/plural-cli/cmd/command/cd"

	"github.com/fatih/color"
	"github.com/urfave/cli"
	"helm.sh/helm/v3/pkg/action"

	"github.com/pluralsh/plural-cli/cmd/command/clone"
	cryptocmd "github.com/pluralsh/plural-cli/cmd/command/crypto"
	"github.com/pluralsh/plural-cli/cmd/command/down"
	cmdinit "github.com/pluralsh/plural-cli/cmd/command/init"
	"github.com/pluralsh/plural-cli/cmd/command/mgmt"
	"github.com/pluralsh/plural-cli/cmd/command/pr"
	"github.com/pluralsh/plural-cli/cmd/command/profile"
	"github.com/pluralsh/plural-cli/cmd/command/up"
	"github.com/pluralsh/plural-cli/cmd/command/version"
	"github.com/pluralsh/plural-cli/pkg/client"
	conf "github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/crypto"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

const ApplicationName = "plrlctl"

type Plural struct {
	client.Plural
	HelmConfiguration *action.Configuration
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
		clone.Command(),
		cd.Command(plural.Plural, plural.HelmConfiguration),
		up.Command(plural.Plural),
		down.Command(),
		pr.Command(plural.Plural),
		cmdinit.Command(plural.Plural),
		mgmt.Command(plural.Plural),
		profile.Command(),
		version.Command(),
	}
	app.Commands = commands
	if os.Getenv("ENABLE_COLOR") != "" {
		color.NoColor = false
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
