package plural

import (
	"os"

	"github.com/urfave/cli"
	"helm.sh/helm/v3/pkg/action"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/crypto"
	"github.com/pluralsh/plural/pkg/kubernetes"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils"
)

func init() {
	cli.BashCompletionFlag = cli.BoolFlag{Name: "compgen", Hidden: true}
}

const ApplicationName = "plural"

type Plural struct {
	api.Client
	kubernetes.Kube
	HelmConfiguration *action.Configuration
}

func (p *Plural) InitKube() error {
	if p.Kube == nil {
		kube, err := kubernetes.Kubernetes()
		if err != nil {
			return err
		}
		p.Kube = kube
	}
	return nil
}

func (p *Plural) InitPluralClient() {
	if p.Client == nil {
		if project, err := manifest.FetchProject(); err == nil && config.Exists() {
			conf := config.Read()
			if owner := project.Owner; owner != nil && conf.Email != owner.Email {
				utils.LogInfo().Printf("Trying to impersonate service account: %s \n", owner.Email)
				jwt, email, err := api.FromConfig(&conf).ImpersonateServiceAccount(owner.Email)
				if err != nil {
					utils.Error("You (%s) are not the owner of this repo %s, %v \n", conf.Email, owner.Email, api.GetErrorResponse(err, "ImpersonateServiceAccount"))
					os.Exit(1)
				}
				conf.Email = email
				conf.Token = jwt
				p.Client = api.FromConfig(&conf)
				accessToken, err := p.Client.GrabAccessToken()
				if err != nil {
					utils.Error("failed to create access token, bailing")
					os.Exit(1)
				}
				conf.Token = accessToken
				config.SetConfig(&conf)
				return
			}
		}

		p.Client = api.NewClient()
	}
}

func (p *Plural) getCommands() []cli.Command {
	return []cli.Command{
		{
			Name:    "version",
			Aliases: []string{"v", "vsn"},
			Usage:   "Gets cli version info",
			Action:  versionInfo,
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
			Action: tracked(rooted(latestVersion(owned(upstreamSynced(p.build)))), "cli.build"),
		},
		{
			Name:      "info",
			Usage:     "Get information for your installation of APP",
			ArgsUsage: "APP",
			Action:    latestVersion(owned(rooted(p.info))),
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
			Action: tracked(latestVersion(owned(rooted(p.deploy))), "cli.deploy"),
		},
		{
			Name:      "diff",
			Aliases:   []string{"df"},
			Usage:     "diffs the state of the current workspace with the deployed version and dumps results to diffs/",
			ArgsUsage: "WKSPACE",
			Action:    latestVersion(handleDiff),
		},
		{
			Name:     "create",
			Usage:    "scaffolds the resources needed to create a new plural repository",
			Action:   latestVersion(handleScaffold),
			Category: "WKSPACE",
		},
		{
			Name:      "watch",
			Usage:     "watches applications until they become ready",
			ArgsUsage: "REPO",
			Action:    latestVersion(initKubeconfig(requireArgs(handleWatch, []string{"REPO"}))),
			Category:  "Debugging",
		},
		{
			Name:      "wait",
			Usage:     "waits on applications until they become ready",
			ArgsUsage: "REPO",
			Action:    latestVersion(requireArgs(handleWait, []string{"REPO"})),
			Category:  "Debugging",
		},
		{
			Name:      "info",
			Usage:     "generates a console dashboard for the namespace of this repo",
			ArgsUsage: "REPO",
			Action:    latestVersion(requireArgs(handleInfo, []string{"REPO"})),
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
			Action:   latestVersion(apply),
			Category: "Publishing",
		},
		{
			Name:     "topsort",
			Aliases:  []string{"top"},
			Usage:    "renders a dependency-inferred topological sort of the installations in a workspace",
			Action:   latestVersion(p.topsort),
			Category: "Workspace",
		},
		{
			Name:     "dependencies",
			Aliases:  []string{"deps"},
			Usage:    "prints ordered dependencies for a repo in your workspace",
			Action:   latestVersion(p.dependencies),
			Category: "Workspace",
		},
		{
			Name:      "bounce",
			Aliases:   []string{"b"},
			Usage:     "redeploys the charts in a workspace",
			ArgsUsage: "WKSPACE",
			Action:    latestVersion(initKubeconfig(owned(p.bounce))),
		},
		{
			Name:      "readme",
			Aliases:   []string{"b"},
			Usage:     "generates the readme for your installation repo",
			ArgsUsage: "WKSPACE",
			Action:    latestVersion(downloadReadme),
		},
		{
			Name:      "destroy",
			Aliases:   []string{"d"},
			Usage:     "iterates through all installations in reverse topological order, deleting helm installations and terraform",
			ArgsUsage: "WKSPACE",
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
			Action: tracked(latestVersion(owned(upstreamSynced(p.destroy))), "cli.destroy"),
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
			Action: latestVersion(requireArgs(p.handleUpgrade, []string{"QUEUE", "REPO"})),
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
			Action: tracked(latestVersion(p.handleInit), "cli.init"),
		},
		{
			Name:   "preflights",
			Usage:  "runs provider preflight checks",
			Action: latestVersion(preflights),
		},
		{
			Name:   "login",
			Usage:  "logs into plural and saves credentials to the current config profile",
			Action: latestVersion(handleLogin),
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
			Action:   latestVersion(handleImport),
			Category: "User Profile",
		},
		{
			Name:     "repair",
			Usage:    "commits any new encrypted changes in your local workspace automatically",
			Action:   latestVersion(handleRepair),
			Category: "Workspace",
		},
		{
			Name:     "serve",
			Usage:    "launch the server",
			Action:   latestVersion(handleServe),
			Category: "Workspace",
		},
		{
			Name:        "shell",
			Usage:       "manages your cloud shell",
			Subcommands: shellCommands(),
			Category:    "Workspace",
		},
		{
			Name:        "clusters",
			Usage:       "commands related to managing plural clusters",
			Subcommands: p.clusterCommands(),
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
			Action:   latestVersion(testTemplate),
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
			Name:        "api",
			Usage:       "inspect the plural api",
			Subcommands: p.apiCommands(),
			Category:    "API",
		},
		{
			Name:        "config",
			Aliases:     []string{"conf"},
			Usage:       "reads/modifies cli configuration",
			Subcommands: configCommands(),
			Category:    "User Profile",
		},
		{
			Name:        "workspace",
			Aliases:     []string{"wkspace"},
			Usage:       "Commands for managing installations in your workspace",
			Subcommands: p.workspaceCommands(),
			Category:    "Workspace",
		},
		{
			Name:        "profile",
			Usage:       "Commands for managing config profiles for plural",
			Subcommands: profileCommands(),
			Category:    "User Profile",
		},
		{
			Name:        "output",
			Usage:       "Commands for generating outputs from supported tools",
			Subcommands: outputCommands(),
			Category:    "Workspace",
		},
		{
			Name:        "logs",
			Usage:       "Commands for tailing logs for specific apps",
			Subcommands: p.logsCommands(),
			Category:    "Debugging",
		},
		{
			Name:        "bundle",
			Usage:       "Commands for installing and discovering installation bundles",
			Subcommands: p.bundleCommands(),
		},
		{
			Name:        "stack",
			Usage:       "Commands for installing and discovering plural stacks",
			Subcommands: p.stackCommands(),
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
			Name:        "utils",
			Usage:       "useful plural utilities",
			Subcommands: utilsCommands(),
			Category:    "Miscellaneous",
		},
		{
			Name:        "vpn",
			Usage:       "interacting with the plural vpn",
			Subcommands: p.vpnCommands(),
			Category:    "Miscellaneous",
		},
		{
			Name:     "ai",
			Usage:    "utilize openai to get help with your setup",
			Action:   p.aiHelp,
			Category: "Debugging",
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
			Action:   latestVersion(handleHelmTemplate),
			Category: "Publishing",
		},
		{
			Name:     "build-context",
			Usage:    "creates a fresh context.yaml for legacy repos",
			Action:   latestVersion(p.buildContext),
			Category: "Workspace",
		},
		{
			Name:     "changed",
			Usage:    "shows repos with pending changes",
			Action:   latestVersion(diffed),
			Category: "Workspace",
		},
		{
			Name:     "from-grafana",
			Usage:    "imports a grafana dashboard to a plural crd",
			Action:   latestVersion(formatDashboard),
			Category: "Publishing",
		},
		p.uiCommands(),
		{
			Name:        "bootstrap",
			Usage:       "Commands for bootstrapping cluster",
			Subcommands: p.bootstrapCommands(),
			Category:    "Bootstrap",
		},
	}
}

func globalFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "profile-file",
			Usage:       "configure your config.yml profile `FILE`",
			EnvVar:      "PLURAL_PROFILE_FILE",
			Destination: &config.ProfileFile,
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
		},
	}
}

func CreateNewApp(plural *Plural) *cli.App {
	app := cli.NewApp()
	app.Name = ApplicationName
	app.Usage = "Tooling to manage your installed plural applications"
	app.EnableBashCompletion = true
	app.Flags = globalFlags()
	app.Commands = plural.getCommands()
	links := linkCommands()
	app.Commands = append(app.Commands, links...)

	return app
}
