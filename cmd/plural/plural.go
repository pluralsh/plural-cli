package plural

import (
	"github.com/pluralsh/plural-cli/cmd/api"
	"github.com/pluralsh/plural-cli/cmd/auth"
	"github.com/pluralsh/plural-cli/cmd/bootstrap"
	"github.com/pluralsh/plural-cli/cmd/bundle"
	"github.com/pluralsh/plural-cli/cmd/cd"
	"github.com/pluralsh/plural-cli/cmd/config"
	"github.com/pluralsh/plural-cli/cmd/log"
	"github.com/pluralsh/plural-cli/cmd/profile"
	"github.com/pluralsh/plural-cli/cmd/stack"
	"github.com/pluralsh/plural-cli/cmd/vpn"
	"github.com/pluralsh/plural-cli/cmd/workspace"
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
			Name:    "version",
			Aliases: []string{"v", "vsn"},
			Usage:   "Gets cli version info",
			Action:  common.VersionInfo,
		},
		{
			Name:  "up",
			Usage: "sets up your repository and an initial management cluster",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "endpoint",
					Usage: "the endpoint for the plural installation you're working with",
				},
				cli.StringFlag{
					Name:  "service-account",
					Usage: "email for the service account you'd like to use for this workspace",
				},
				cli.BoolFlag{
					Name:  "ignore-preflights",
					Usage: "whether to ignore preflight check failures prior to init",
				},
				cli.StringFlag{
					Name:  "commit",
					Usage: "commits your changes with this message",
				},
			},
			Action: common.LatestVersion(p.handleUp),
		},
		{
			Name:   "down",
			Usage:  "destroys your management cluster and any apps installed on it",
			Action: common.LatestVersion(p.handleDown),
		},
		{
			Name:  "init",
			Usage: "initializes plural within a git repo",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "endpoint",
					Usage: "the endpoint for the plural installation you're working with",
				},
				cli.StringFlag{
					Name:  "service-account",
					Usage: "email for the service account you'd like to use for this workspace",
				},
				cli.BoolFlag{
					Name:  "ignore-preflights",
					Usage: "whether to ignore preflight check failures prior to init",
				},
			},
			Action: tracked(common.LatestVersion(p.handleInit), "cli.init"),
		},
		{
			Name:    "build",
			Aliases: []string{"bld"},
			Usage:   "builds your workspace",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "only",
					Usage: "repository to (re)build",
				},
				cli.BoolFlag{
					Name:  "force",
					Usage: "force workspace to build even if remote is out of sync",
				},
			},
			Action: tracked(rooted(common.LatestVersion(owned(upstreamSynced(p.build)))), "cli.build"),
		},
		{
			Name:      "info",
			Usage:     "Get information for your installation of APP",
			ArgsUsage: "APP",
			Action:    common.LatestVersion(owned(rooted(p.info))),
		},
		{
			Name:      "deploy",
			Aliases:   []string{"d"},
			Usage:     "Deploys the current workspace. This command will first sniff out git diffs in workspaces, topsort them, then apply all changes.",
			ArgsUsage: "Workspace",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "silence",
					Usage: "don't display notes for deployed apps",
				},
				cli.BoolFlag{
					Name:  "verbose",
					Usage: "show all command output during execution",
				},
				cli.BoolFlag{
					Name:  "ignore-console",
					Usage: "don't deploy the plural console",
				},
				cli.BoolFlag{
					Name:  "all",
					Usage: "deploy all repos irregardless of changes",
				},
				cli.StringFlag{
					Name:  "commit",
					Usage: "commits your changes with this message",
				},
				cli.StringSliceFlag{
					Name:  "from",
					Usage: "deploys only this application and its dependencies",
				},
				cli.BoolFlag{
					Name:  "force",
					Usage: "use force push when pushing to git",
				},
			},
			Action: tracked(common.LatestVersion(owned(rooted(p.deploy))), "cli.deploy"),
		},
		{
			Name:      "diff",
			Aliases:   []string{"df"},
			Usage:     "diffs the state of the current workspace with the deployed version and dumps results to diffs/",
			ArgsUsage: "APP",
			Action:    common.LatestVersion(handleDiff),
		},
		{
			Name:      "clone",
			Usage:     "clones and decrypts a plural repo",
			ArgsUsage: "URL",
			Action:    handleClone,
		},
		{
			Name:     "create",
			Usage:    "scaffolds the resources needed to create a new plural repository",
			Action:   common.LatestVersion(handleScaffold),
			Category: "Workspace",
		},
		{
			Name:      "watch",
			Usage:     "watches applications until they become ready",
			ArgsUsage: "REPO",
			Action:    common.LatestVersion(initKubeconfig(requireArgs(handleWatch, []string{"REPO"}))),
			Category:  "Debugging",
		},
		{
			Name:      "wait",
			Usage:     "waits on applications until they become ready",
			ArgsUsage: "REPO",
			Action:    common.LatestVersion(requireArgs(handleWait, []string{"REPO"})),
			Category:  "Debugging",
		},
		{
			Name:      "info",
			Usage:     "generates a console dashboard for the namespace of this repo",
			ArgsUsage: "REPO",
			Action:    common.LatestVersion(requireArgs(handleInfo, []string{"REPO"})),
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
			Action:   common.LatestVersion(apply),
			Category: "Publishing",
		},
		{
			Name:      "bounce",
			Aliases:   []string{"b"},
			Usage:     "redeploys the charts in a workspace",
			ArgsUsage: "APP",
			Action:    common.LatestVersion(initKubeconfig(owned(p.bounce))),
		},
		{
			Name:     "readme",
			Aliases:  []string{"b"},
			Usage:    "generates the readme for your installation repo",
			Category: "Workspace",
			Action:   common.LatestVersion(downloadReadme),
		},
		{
			Name:      "destroy",
			Aliases:   []string{"d"},
			Usage:     "iterates through all installations in reverse topological order, deleting helm installations and terraform",
			ArgsUsage: "APP",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "from",
					Usage: "where to start your deploy command (useful when restarting interrupted destroys)",
				},
				cli.StringFlag{
					Name:  "commit",
					Usage: "commits your changes with this message",
				},
				cli.BoolFlag{
					Name:  "force",
					Usage: "use force push when pushing to git",
				},
				cli.BoolFlag{
					Name:  "all",
					Usage: "tear down the entire cluster gracefully in one go",
				},
			},
			Action: tracked(common.LatestVersion(owned(upstreamSynced(p.destroy))), "cli.destroy"),
		},
		{
			Name:      "upgrade",
			Usage:     "creates an upgrade in the upgrade queue QUEUE for application REPO",
			ArgsUsage: "QUEUE REPO",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "f",
					Usage: "file containing upgrade contents, use - for stdin",
				},
			},
			Action: common.LatestVersion(requireArgs(p.handleUpgrade, []string{"QUEUE", "REPO"})),
		},
		{
			Name:     "preflights",
			Usage:    "runs provider preflight checks",
			Category: "Workspace",
			Action:   common.LatestVersion(preflights),
		},
		{
			Name:   "login",
			Usage:  "logs into plural and saves credentials to the current config profile",
			Action: common.LatestVersion(handleLogin),
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
			Action:   common.LatestVersion(handleImport),
			Category: "User Profile",
		},
		{
			Name:     "repair",
			Usage:    "commits any new encrypted changes in your local workspace automatically",
			Action:   common.LatestVersion(handleRepair),
			Category: "Workspace",
		},
		{
			Name:     "serve",
			Usage:    "launch the server",
			Action:   common.LatestVersion(handleServe),
			Category: "Workspace",
		},
		{
			Name:        "shell",
			Usage:       "manages your cloud shell",
			Subcommands: shellCommands(),
			Category:    "Workspace",
		},
		{
			Name:        "repos",
			Usage:       "view and manage plural repositories",
			Subcommands: p.reposCommands(),
			Category:    "API",
		},
		{
			Name:        "apps",
			Usage:       "view and manage plural repositories",
			Subcommands: p.reposCommands(),
			Category:    "API",
		},
		{
			Name:     "test",
			Usage:    "validate a values templace",
			Action:   common.LatestVersion(testTemplate),
			Category: "Publishing",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "templateType",
					Usage: "Determines the template type. Go template by default",
				},
			},
		},
		{
			Name:        "proxy",
			Usage:       "proxies into running processes in your cluster",
			Subcommands: p.proxyCommands(),
			Category:    "Debugging",
		},
		{
			Name:        "crypto",
			Usage:       "plural encryption utilities",
			Subcommands: p.cryptoCommands(),
			Category:    "User Profile",
		},
		{
			Name:        "push",
			Usage:       "utilities for pushing tf or helm packages",
			Subcommands: p.pushCommands(),
			Category:    "Publishing",
		},
		{
			Name:        "output",
			Usage:       "Commands for generating outputs from supported tools",
			Subcommands: outputCommands(),
			Category:    "Workspace",
		},
		{
			Name:        "packages",
			Usage:       "Commands for managing your installed packages",
			Subcommands: p.packagesCommands(),
		},
		{
			Name:        "ops",
			Usage:       "Commands for simplifying cluster operations",
			Subcommands: p.opsCommands(),
			Category:    "Debugging",
		},
		{
			Name:     "ai",
			Usage:    "utilize openai to get help with your setup",
			Action:   p.aiHelp,
			Category: "Debugging",
		},
		{
			Name:        "pull-requests",
			Aliases:     []string{"pr"},
			Usage:       "Generate and manage pull requests",
			Subcommands: prCommands(),
			Category:    "CD",
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
			Action:   common.LatestVersion(handleHelmTemplate),
			Category: "Publishing",
		},
		{
			Name:     "changed",
			Usage:    "shows repos with pending changes",
			Action:   common.LatestVersion(diffed),
			Category: "Workspace",
		},
		p.uiCommands(),
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
			Destination: &bootstrapMode,
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
		bootstrap.Command(plural.Plural),
		bundle.Command(plural.Plural),
		cd.Command(plural.Plural, plural.HelmConfiguration),
		config.Command(),
		profile.Command(),
		stack.Command(plural.Plural),
		log.Command(plural.Plural),
		workspace.Command(plural.Plural, plural.HelmConfiguration),
		vpn.Command(plural.Plural),
	}
	commands = append(commands, plural.getCommands()...)
	app.Commands = commands
	links := linkCommands()
	app.Commands = append(app.Commands, links...)

	return app
}

func RunPlural(arguments []string) error {
	return CreateNewApp(&Plural{}).Run(arguments)
}
