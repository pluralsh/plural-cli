package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/urfave/cli"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	app := cli.NewApp()
	app.Name = "plural"
	app.Usage = "Tooling to manage your installed plural applications"

	app.Commands = []cli.Command{
		{
			Name:    "build",
			Aliases: []string{"b"},
			Usage:   "builds your workspace",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "only",
					Usage: "repository to (re)build",
				},
			},
			Action: build,
		},
		{
			Name:      "deploy",
			Aliases:   []string{"d"},
			Usage:     "deploys the current workspace",
			ArgsUsage: "WKSPACE",
			Action:    deploy,
		},
		{
			Name:      "diff",
			Aliases:   []string{"df"},
			Usage:     "diffs the state of  the current workspace with the deployed version and dumps results to diffs/",
			ArgsUsage: "WKSPACE",
			Action:    handleDiff,
		},
		{
			Name:      "watch",
			Usage:     "watches applications until they become ready",
			ArgsUsage: "REPO",
			Action:    handleWatch,
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
			Action:   apply,
			Category: "Publishing",
		},
		{
			Name:    "validate",
			Aliases: []string{"v"},
			Usage:   "validates your workspace",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "only",
					Usage: "repository to (re)build",
				},
			},
			Action:   validate,
			Category: "Workspace",
		},
		{
			Name:     "topsort",
			Aliases:  []string{"d"},
			Usage:    "renders a dependency-inferred topological sort of the installations in a workspace",
			Action:   topsort,
			Category: "Workspace",
		},
		{
			Name:      "bounce",
			Aliases:   []string{"b"},
			Usage:     "redeploys the charts in a workspace",
			ArgsUsage: "WKSPACE",
			Action:    bounce,
		},
		{
			Name:      "destroy",
			Aliases:   []string{"b"},
			Usage:     "iterates through all installations in reverse topological order, deleting helm installations and terraform",
			ArgsUsage: "WKSPACE",
			Action:    destroy,
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
			},
			Action: handleInit,
		},
		{
			Name:   "login",
			Usage:  "logs into plural and saves credentials to the current config profile",
			Action: handleLogin,
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
			Name:     "install",
			Usage:    "installs forge cli dependencies",
			Action:   handleInstall,
			Category: "Workspace",
		},
		{
			Name:     "import",
			Usage:    "imports forge config from another file",
			Action:   handleImport,
			Category: "User Profile",
		},
		{
			Name:     "test",
			Usage:    "validate a values templace",
			Action:   testTemplate,
			Category: "Publishing",
		},
		{
			Name:        "proxy",
			Usage:       "proxies into running processes in your cluster",
			Subcommands: proxyCommands(),
			Category:    "Debugging",
		},
		{
			Name:        "crypto",
			Usage:       "forge encryption utilities",
			Subcommands: cryptoCommands(),
			Category:    "User Profile",
		},
		{
			Name:        "push",
			Usage:       "utilities for pushing tf or helm packages",
			Subcommands: pushCommands(),
			Category:    "Publishing",
		},
		{
			Name:        "api",
			Usage:       "inspect the forge api",
			Subcommands: apiCommands(),
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
			Subcommands: workspaceCommands(),
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
			Subcommands: logsCommands(),
			Category:    "Debugging",
		},
		{
			Name:        "bundle",
			Usage:       "Commands for installing and discovering installation bundles",
			Subcommands: bundleCommands(),
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
			Action:   handleHelmTemplate,
			Category: "Publishing",
		},
		{
			Name:    "upgrade",
			Aliases: []string{"up"},
			Usage:   "Creates an upgrade for a repository",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name",
					Usage: "the repository name",
				},
				cli.StringFlag{
					Name:  "message",
					Usage: "a message describing the upgrade",
				},
			},
			Action:   handleUpgrade,
			Category: "API",
		},
		{
			Name:     "build-context",
			Usage:    "creates a fresh context.yaml for legacy repos",
			Action:   buildContext,
			Category: "Workspace",
		},
		{
			Name:     "changed",
			Usage:    "shows repos with pending changes",
			Action:   diffed,
			Category: "Workspace",
		},
		{
			Name:     "from-grafana",
			Usage:    "imports a grafana dashboard to a plural crd",
			Action:   formatDashboard,
			Category: "Publishing",
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
