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
			Name:  "apply",
			Usage: "applys the current pluralfile",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "file, f",
					Usage: "pluralfile to use",
				},
			},
			Action: apply,
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
			Action: validate,
		},
		{
			Name:    "topsort",
			Aliases: []string{"d"},
			Usage:   "renders a dependency-inferred topological sort of the installations in a workspace",
			Action:  topsort,
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
			Name:   "init",
			Usage:  "initializes plural within a git repo",
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
		},
		{
			Name:   "install",
			Usage:  "installs forge cli dependencies",
			Action: handleInstall,
		},
		{
			Name:   "import",
			Usage:  "imports forge config from another file",
			Action: handleImport,
		},
		{
			Name:   "test",
			Usage:  "validate a values templace",
			Action: testTemplate,
		},
		{
			Name:        "proxy",
			Usage:       "proxies into running processes in your cluster",
			Subcommands: proxyCommands(),
		},
		{
			Name:        "crypto",
			Usage:       "forge encryption utilities",
			Subcommands: cryptoCommands(),
		},
		{
			Name:        "push",
			Usage:       "utilities for pushing tf or helm packages",
			Subcommands: pushCommands(),
		},
		{
			Name:        "api",
			Usage:       "inspect the forge api",
			Subcommands: apiCommands(),
		},
		{
			Name:        "config",
			Aliases:     []string{"conf"},
			Usage:       "reads/modifies cli configuration",
			Subcommands: configCommands(),
		},
		{
			Name:        "workspace",
			Aliases:     []string{"wkspace"},
			Usage:       "Commands for managing installations in your workspace",
			Subcommands: workspaceCommands(),
		},
		{
			Name:        "profile",
			Usage:       "Commands for managing config profiles for plural",
			Subcommands: profileCommands(),
		},
		{
			Name:        "output",
			Usage:       "Commands for generating outputs from supported tools",
			Subcommands: outputCommands(),
		},
		{
			Name:        "logs",
			Usage:       "Commands for tailing logs for specific apps",
			Subcommands: logsCommands(),
		},
		{
			Name: "template",
			Aliases: []string{"tpl"},
			Usage: "templates a helm chart to be uploaded to plural",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "values",
					Usage: "the values file",
				},
			},
			Action: handleHelmTemplate,
		},
		{
			Name: "upgrade",
			Aliases: []string{"up"},
			Usage: "Creates an upgrade for a repository",
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
			Action: handleUpgrade,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
