package plural

import (
	"github.com/pluralsh/plural-cli/cmd/command/api"
	"github.com/pluralsh/plural-cli/cmd/command/auth"
	"github.com/pluralsh/plural-cli/cmd/command/cd"
	"github.com/pluralsh/plural-cli/cmd/command/clone"
	"github.com/pluralsh/plural-cli/cmd/command/config"
	cryptocmd "github.com/pluralsh/plural-cli/cmd/command/crypto"
	"github.com/pluralsh/plural-cli/cmd/command/down"
	cmdinit "github.com/pluralsh/plural-cli/cmd/command/init"
	"github.com/pluralsh/plural-cli/cmd/command/mgmt"
	"github.com/pluralsh/plural-cli/cmd/command/pr"
	"github.com/pluralsh/plural-cli/cmd/command/profile"
	"github.com/pluralsh/plural-cli/cmd/command/stacks"
	"github.com/pluralsh/plural-cli/cmd/command/up"
	"github.com/pluralsh/plural-cli/cmd/command/version"
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	conf "github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/crypto"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/urfave/cli"
	"helm.sh/helm/v3/pkg/action"
)

func init() {
	cli.BashCompletionFlag = cli.BoolFlag{Name: "compgen", Hidden: true}
}

const ApplicationName = "plural"

type Plural struct {
	client.Plural
	HelmConfiguration *action.Configuration
}

func (p *Plural) getCommands() []cli.Command {
	return []cli.Command{
		{
			Name:     "preflights",
			Usage:    "runs provider preflight checks",
			Category: "Workspace",
			Action:   common.LatestVersion(common.Preflights),
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
		{
			Name:     "import",
			Usage:    "imports plural config from another file",
			Action:   common.LatestVersion(common.HandleImport),
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

func CreateNewApp(plural *Plural) *cli.App {
	if plural == nil {
		plural = &Plural{}
	}
	app := cli.NewApp()
	app.Name = ApplicationName
	app.Usage = "Tooling to manage your installed plural applications"
	app.EnableBashCompletion = true
	app.Flags = globalFlags()
	commands := []cli.Command{
		api.Command(plural.Plural),
		auth.Command(plural.Plural),
		cd.Command(plural.Plural, plural.HelmConfiguration),
		config.Command(),
		cryptocmd.Command(plural.Plural),
		clone.Command(),
		down.Command(),
		mgmt.Command(plural.Plural),
		profile.Command(),
		stacks.Command(plural.Plural),
		pr.Command(plural.Plural),
		cmdinit.Command(plural.Plural),
		up.Command(plural.Plural),
		version.Command(),
	}
	commands = append(commands, plural.getCommands()...)
	app.Commands = commands

	return app
}
