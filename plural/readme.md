```bash
NAME:
   plural - Tooling to manage your installed plural applications

USAGE:
   plural [global options] command [command options] [arguments...]

COMMANDS:
   build, b    builds your workspace
   deploy, d   deploys the current workspace
   diff, df    diffs the state of  the current workspace with the deployed version and dumps results to diffs/
   bounce, b   redeploys the charts in a workspace
   destroy, b  iterates through all installations in reverse topological order, deleting helm installations and terraform
   init        initializes plural within a git repo
   bundle      Commands for installing and discovering installation bundles
   help, h     Shows a list of commands or help for one command

   API:
     api          inspect the forge api
     upgrade, up  Creates an upgrade for a repository

   Debugging:
     watch  watches applications until they become ready
     proxy  proxies into running processes in your cluster
     logs   Commands for tailing logs for specific apps

   Publishing:
     apply          applys the current pluralfile
     test           validate a values templace
     push           utilities for pushing tf or helm packages
     template, tpl  templates a helm chart to be uploaded to plural
     from-grafana   imports a grafana dashboard to a plural crd

   User Profile:
     login         logs into plural and saves credentials to the current config profile
     import        imports forge config from another file
     crypto        forge encryption utilities
     config, conf  reads/modifies cli configuration
     profile       Commands for managing config profiles for plural

   Workspace:
     validate, v         validates your workspace
     topsort, d          renders a dependency-inferred topological sort of the installations in a workspace
     install             installs forge cli dependencies
     workspace, wkspace  Commands for managing installations in your workspace
     output              Commands for generating outputs from supported tools
     build-context       creates a fresh context.yaml for legacy repos
     changed             shows repos with pending changes

GLOBAL OPTIONS:
   --help, -h  show help
```
