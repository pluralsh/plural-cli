package plural

import (
	"github.com/pluralsh/plural-cli/cmd/command/ai"
	"github.com/pluralsh/plural-cli/cmd/command/api"
	"github.com/pluralsh/plural-cli/cmd/command/auth"
	"github.com/pluralsh/plural-cli/cmd/command/cd"
	"github.com/pluralsh/plural-cli/cmd/command/clone"
	"github.com/pluralsh/plural-cli/cmd/command/clusters"
	"github.com/pluralsh/plural-cli/cmd/command/config"
	cryptocmd "github.com/pluralsh/plural-cli/cmd/command/crypto"
	"github.com/pluralsh/plural-cli/cmd/command/down"
	cmdinit "github.com/pluralsh/plural-cli/cmd/command/init"
	"github.com/pluralsh/plural-cli/cmd/command/log"
	"github.com/pluralsh/plural-cli/cmd/command/ops"
	"github.com/pluralsh/plural-cli/cmd/command/pr"
	"github.com/pluralsh/plural-cli/cmd/command/profile"
	"github.com/pluralsh/plural-cli/cmd/command/proxy"
	"github.com/pluralsh/plural-cli/cmd/command/up"
	"github.com/pluralsh/plural-cli/cmd/command/version"
	"github.com/pluralsh/plural-cli/cmd/command/vpn"
	"github.com/pluralsh/plural-cli/cmd/command/workspace"
	"github.com/pluralsh/plural-cli/pkg/client"
	"github.com/pluralsh/plural-cli/pkg/common"
	conf "github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/crypto"
	"github.com/pluralsh/plural-cli/pkg/exp"
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
			Name:      "diff",
			Aliases:   []string{"df"},
			Usage:     "diffs the state of the current workspace with the deployed version and dumps results to diffs/",
			ArgsUsage: "APP",
			Action:    common.LatestVersion(common.HandleDiff),
		},
		{
			Name:     "create",
			Usage:    "scaffolds the resources needed to create a new plural repository",
			Action:   common.LatestVersion(common.HandleScaffold),
			Category: "Workspace",
		},
		{
			Name:      "watch",
			Usage:     "watches applications until they become ready",
			ArgsUsage: "REPO",
			Action:    common.LatestVersion(common.InitKubeconfig(common.RequireArgs(common.HandleWatch, []string{"REPO"}))),
			Category:  "Debugging",
		},
		{
			Name:      "wait",
			Usage:     "waits on applications until they become ready",
			ArgsUsage: "REPO",
			Action:    common.LatestVersion(common.RequireArgs(common.HandleWait, []string{"REPO"})),
			Category:  "Debugging",
		},
		{
			Name:      "info",
			Usage:     "generates a console dashboard for the namespace of this repo",
			ArgsUsage: "REPO",
			Action:    common.LatestVersion(common.RequireArgs(common.HandleInfo, []string{"REPO"})),
			Category:  "Debugging",
		},
		{
			Name:  "apply",
			Usage: "applys the current pluralfile",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "file, f",
					Usage: "pluralfile to use",
				},
			},
			Action:   common.LatestVersion(common.Apply),
			Category: "Publishing",
		},

		{
			Name:     "readme",
			Aliases:  []string{"b"},
			Usage:    "generates the readme for your installation repo",
			Category: "Workspace",
			Action:   common.LatestVersion(common.DownloadReadme),
		},
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
		{
			Name:     "repair",
			Usage:    "commits any new encrypted changes in your local workspace automatically",
			Action:   common.LatestVersion(common.HandleRepair),
			Category: "Workspace",
		},
		{
			Name:     "serve",
			Usage:    "launch the server",
			Action:   common.LatestVersion(common.HandleServe),
			Category: "Workspace",
		},
		{
			Name:     "test",
			Usage:    "validate a values templace",
			Action:   common.LatestVersion(common.TestTemplate),
			Category: "Publishing",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "templateType",
					Usage: "Determines the template type. Go template by default",
				},
			},
		},
		{
			Name:    "template",
			Aliases: []string{"tpl"},
			Usage:   "templates a helm chart to be uploaded to plural",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "values",
					Usage: "the values file",
				},
			},
			Action:   common.LatestVersion(common.HandleHelmTemplate),
			Category: "Publishing",
		},
		{
			Name:     "changed",
			Usage:    "shows repos with pending changes",
			Action:   common.LatestVersion(common.Diffed),
			Category: "Workspace",
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
		cli.BoolFlag{
			Name:        "bootstrap",
			Usage:       "enable bootstrap mode",
			Destination: &common.BootstrapMode,
			Hidden:      !exp.IsFeatureEnabled(exp.EXP_PLURAL_CAPI),
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
		ai.Command(plural.Plural),
		cd.Command(plural.Plural, plural.HelmConfiguration),
		config.Command(),
		cryptocmd.Command(plural.Plural),
		clusters.Command(plural.Plural),
		clone.Command(),
		down.Command(),
		ops.Command(plural.Plural),
		profile.Command(),
		pr.Command(plural.Plural),
		cmdinit.Command(plural.Plural),
		up.Command(plural.Plural),
		workspace.Command(plural.Plural, plural.HelmConfiguration),
		vpn.Command(plural.Plural),
		version.Command(),
	}
	commands = append(commands, plural.getCommands()...)
	app.Commands = commands

	return app
}
