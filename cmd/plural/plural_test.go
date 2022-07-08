package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const pluralOutput = `NAME:
   plural - Tooling to manage your installed plural applications

USAGE:
   plural [global options] command [command options] [arguments...]

COMMANDS:
   version, v, vsn  Gets cli version info
   build, b         builds your workspace
   deploy, d        Deploys the current workspace. This command will first sniff out git diffs in workspaces, topsort them, then apply all changes.
   diff, df         diffs the state of the current workspace with the deployed version and dumps results to diffs/
   bounce, b        redeploys the charts in a workspace
   destroy, b       iterates through all installations in reverse topological order, deleting helm installations and terraform
   init             initializes plural within a git repo
   preflights       runs provider preflight checks
   bundle           Commands for installing and discovering installation bundles
   link             links a local package into an installation repo
   unlink           unlinks a linked package
   help, h          Shows a list of commands or help for one command

   API:
     repos        view and manage plural repositories
     api          inspect the forge api
     upgrade, up  Creates an upgrade for a repository

   Debugging:
     watch  watches applications until they become ready
     wait   waits on applications until they become ready
     info   generates a console dashboard for the namespace of this repo
     proxy  proxies into running processes in your cluster
     logs   Commands for tailing logs for specific apps
     ops    Commands for simplifying cluster operations

   Miscellaneous:
     utils  useful plural utilities

   Publishing:
     apply          applys the current pluralfile
     test           validate a values templace
     push           utilities for pushing tf or helm packages
     template, tpl  templates a helm chart to be uploaded to plural
     from-grafana   imports a grafana dashboard to a plural crd

   User Profile:
     login         logs into plural and saves credentials to the current config profile
     import        imports plural config from another file
     crypto        forge encryption utilities
     config, conf  reads/modifies cli configuration
     profile       Commands for managing config profiles for plural

   WKSPACE:
     create  scaffolds the resources needed to create a new plural repository

   WORKSPACE:
     repair  commits any new encrypted changes in your local workspace automatically

   Workspace:
     validate, v         validates your workspace
     topsort, d          renders a dependency-inferred topological sort of the installations in a workspace
     serve               launch the server
     shell               manages your cloud shell
     workspace, wkspace  Commands for managing installations in your workspace
     output              Commands for generating outputs from supported tools
     build-context       creates a fresh context.yaml for legacy repos
     changed             shows repos with pending changes

GLOBAL OPTIONS:
   --help, -h  show help
`

func TestPluralApplication(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedResponse string
	}{
		{
			name:             "test plural CLI without arguments",
			args:             []string{ApplicationName},
			expectedResponse: pluralOutput,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := CreateNewApp()
			app.HelpName = ApplicationName
			writer := &bytes.Buffer{}
			app.Writer = writer

			os.Args = test.args
			err := app.Run(os.Args)
			assert.NoError(t, err)
			response := writer.String()
			assert.Equal(t, response, test.expectedResponse)
		})
	}
}
