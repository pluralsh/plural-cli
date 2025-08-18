package cd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/common"
	"github.com/pluralsh/polly/fs"
	lua "github.com/yuin/gopher-lua"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/cd/template"
	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/polly/containers"
	"github.com/pluralsh/polly/luautils"
	"github.com/samber/lo"
	"github.com/urfave/cli"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const pluralDigestHeader = "x-plrl-digest"

func (p *Plural) cdServices() cli.Command {
	return cli.Command{
		Name:        "services",
		Subcommands: p.cdServiceCommands(),
		Usage:       "manage CD services",
	}
}

func (p *Plural) cdServiceCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "list",
			ArgsUsage: "@{cluster-handle}",
			Action:    common.LatestVersion(common.RequireArgs(p.handleListClusterServices, []string{"@{cluster-handle}"})),
			Usage:     "list cluster services",
		},
		{
			Name:      "create",
			ArgsUsage: "@{cluster-handle}",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "name", Usage: "service name", Required: true},
				cli.StringFlag{Name: "namespace", Usage: "service namespace. If not specified the 'default' will be used"},
				cli.StringFlag{Name: "version", Usage: "service version. If not specified the '0.0.1' will be used"},
				cli.StringFlag{Name: "repo-id", Usage: "repository ID", Required: true},
				cli.StringFlag{Name: "git-ref", Usage: "git ref, can be branch, tag or commit sha", Required: true},
				cli.StringFlag{Name: "git-folder", Usage: "folder within the source tree where manifests are located", Required: true},
				cli.StringFlag{Name: "kustomize-folder", Usage: "folder within the kustomize file is located"},
				cli.BoolFlag{Name: "dry-run", Usage: "dry run mode"},
				cli.StringSliceFlag{
					Name:  "conf",
					Usage: "config name value",
				},
				cli.StringFlag{Name: "config-file", Usage: "path for configuration file"},
			},
			Action: common.LatestVersion(common.RequireArgs(p.handleCreateClusterService, []string{"@{cluster-handle}"})),
			Usage:  "create cluster service",
		},
		{
			Name:      "update",
			ArgsUsage: "{service-id}",
			Action:    common.LatestVersion(common.RequireArgs(p.handleUpdateClusterService, []string{"{service-id}"})),
			Usage:     "update cluster service",
			Flags: []cli.Flag{
				cli.StringFlag{Name: "version", Usage: "service version"},
				cli.StringFlag{Name: "git-ref", Usage: "git ref, can be branch, tag or commit sha"},
				cli.StringFlag{Name: "git-folder", Usage: "folder within the source tree where manifests are located"},
				cli.StringFlag{Name: "kustomize-folder", Usage: "folder within the kustomize file is located"},
				cli.StringSliceFlag{
					Name:  "conf",
					Usage: "config name value",
				},
				cli.BoolFlag{Name: "dry-run", Usage: "dry run mode"},
				cli.BoolFlag{Name: "templated", Usage: "set templated flag"},
				cli.StringSliceFlag{
					Name:  "context-id",
					Usage: "bind service context",
				},
			},
		},
		{
			Name:      "clone",
			ArgsUsage: "@{cluster-handle} @{cluster-handle}/{serviceName}",
			Action: common.LatestVersion(common.RequireArgs(p.handleCloneClusterService,
				[]string{"@{cluster-handle}", "@{cluster-handle}/{serviceName}"})),
			Flags: []cli.Flag{
				cli.StringFlag{Name: "name", Usage: "the name for the cloned service", Required: true},
				cli.StringFlag{Name: "namespace", Usage: "the namespace for this cloned service", Required: true},
				cli.StringSliceFlag{
					Name:  "conf",
					Usage: "config name value",
				},
			},
			Usage: "deep clone a service onto either the same cluster or another",
		},
		{
			Name:      "describe",
			ArgsUsage: "@{cluster-handle}/{serviceName}",
			Action:    common.LatestVersion(common.RequireArgs(p.handleDescribeClusterService, []string{"@{cluster-handle}/{serviceName}"})),
			Flags:     []cli.Flag{cli.StringFlag{Name: "o", Usage: "output format"}},
			Usage:     "describe cluster service",
		},
		{
			Name:   "template",
			Action: p.handleTemplateService,
			Usage:  "Dry-runs templating a .liquid or .tpl file with either a full service as params or custom config",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "service",
					Usage: "specify the service you want to use as context while templating"},
				cli.StringFlag{
					Name:  "configuration",
					Usage: "hand-coded configuration for templating (useful if you want to test before creating a service)",
				},
				cli.StringFlag{
					Name:  "file",
					Usage: "The .liquid or .tpl file you want to attempt to template.",
				},
			},
		},
		{
			Name:   "lua",
			Action: p.handleLuaTemplate,
			Usage:  "Templates a .lua file using the Plural defined lua engine and returns the result",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "lua-file",
					Usage: "The .lua file you want to attempt to template.",
				},
				cli.StringFlag{
					Name:  "context",
					Usage: "A yaml context file to imitate the internal service template context",
				},
				cli.StringFlag{
					Name:  "dir",
					Usage: "The directory to run the lua script from, defaults to the current working directory",
				},
			},
		},
		{
			Name:      "delete",
			ArgsUsage: "@{cluster-handle}/{serviceName}",
			Action:    common.LatestVersion(common.RequireArgs(p.handleDeleteClusterService, []string{"@{cluster-handle}/{serviceName}"})),
			Usage:     "delete cluster service",
		},
		{
			Name:      "kick",
			ArgsUsage: "@{cluster-handle}/{serviceName}",
			Action:    common.LatestVersion(common.RequireArgs(p.handleKickClusterService, []string{"@{cluster-handle}/{serviceName}"})),
			Usage:     "force sync cluster service",
		},
		{
			Name:      "tarball",
			ArgsUsage: "@{cluster-handle}/{serviceName}",
			Action:    common.LatestVersion(common.RequireArgs(p.handleTarballClusterService, []string{"@{cluster-handle}/{serviceName}"})),
			Usage:     "download service tarball locally",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dir",
					Usage: "directory to download to",
					Value: ".",
				},
			},
		},
	}
}

func (p *Plural) handleListClusterServices(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	sd, err := p.ConsoleClient.ListClusterServices(common.GetIdAndName(c.Args().Get(0)))
	if err != nil {
		return err
	}
	if sd == nil {
		return fmt.Errorf("returned objects list [ListClusterServices] is nil")
	}
	headers := []string{"Id", "Name", "Namespace", "Git Ref", "Git Folder", "Repo"}
	return utils.PrintTable(sd, headers, func(sd *gqlclient.ServiceDeploymentEdgeFragment) ([]string, error) {
		ref := ""
		folder := ""
		url := ""
		if sd.Node.Git != nil {
			ref = sd.Node.Git.Ref
			folder = sd.Node.Git.Folder
		}
		if sd.Node.Repository != nil {
			url = sd.Node.Repository.URL
		}
		return []string{sd.Node.ID, sd.Node.Name, sd.Node.Namespace, ref, folder, url}, nil
	})
}

func (p *Plural) handleCreateClusterService(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	v, err := validateFlag(c, "version", "0.0.1")
	if err != nil {
		return err
	}
	name := c.String("name")
	namespace, err := validateFlag(c, "namespace", "default")
	if err != nil {
		return err
	}
	repoId := c.String("repo-id")
	gitRef := c.String("git-ref")
	gitFolder := c.String("git-folder")
	dryRun := c.Bool("dry-run")
	attributes := gqlclient.ServiceDeploymentAttributes{
		Name:         name,
		Namespace:    namespace,
		Version:      &v,
		RepositoryID: lo.ToPtr(repoId),
		Git: &gqlclient.GitRefAttributes{
			Ref:    gitRef,
			Folder: gitFolder,
		},
		Configuration: []*gqlclient.ConfigAttributes{},
		DryRun:        lo.ToPtr(dryRun),
	}

	if c.String("kustomize-folder") != "" {
		attributes.Kustomize = &gqlclient.KustomizeAttributes{
			Path: c.String("kustomize-folder"),
		}
	}

	if c.String("config-file") != "" {
		configFile, err := utils.ReadFile(c.String("config-file"))
		if err != nil {
			return err
		}
		sdc := ServiceDeploymentAttributesConfiguration{}
		if err := yaml.Unmarshal([]byte(configFile), &sdc); err != nil {
			return err
		}
		attributes.Configuration = append(attributes.Configuration, sdc.Configuration...)
	}
	var confArgs []string
	if c.IsSet("conf") {
		confArgs = append(confArgs, c.StringSlice("conf")...)
	}
	for _, conf := range confArgs {
		configurationPair := strings.Split(conf, "=")
		if len(configurationPair) == 2 {
			attributes.Configuration = append(attributes.Configuration, &gqlclient.ConfigAttributes{
				Name:  configurationPair[0],
				Value: &configurationPair[1],
			})
		}
	}

	clusterId, clusterName := common.GetIdAndName(c.Args().Get(0))
	sd, err := p.ConsoleClient.CreateClusterService(clusterId, clusterName, attributes)
	if err != nil {
		return err
	}
	if sd == nil {
		return fmt.Errorf("the returned object is empty, check if all fields are set")
	}

	headers := []string{"Id", "Name", "Namespace", "Git Ref", "Git Folder", "Repo"}
	return utils.PrintTable([]*gqlclient.ServiceDeploymentExtended{sd}, headers, func(sd *gqlclient.ServiceDeploymentExtended) ([]string, error) {
		return []string{sd.ID, sd.Name, sd.Namespace, sd.Git.Ref, sd.Git.Folder, sd.Repository.URL}, nil
	})
}

func (p *Plural) handleTemplateService(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	printResult := func(out []byte) error {
		fmt.Println()
		fmt.Println(string(out))
		return nil
	}

	if identifier := c.String("service"); identifier != "" {
		serviceId, clusterName, serviceName, err := getServiceIdClusterNameServiceName(identifier)
		if err != nil {
			return err
		}

		existing, err := p.ConsoleClient.GetClusterService(serviceId, serviceName, clusterName)
		if err != nil {
			return err
		}
		if existing == nil {
			return fmt.Errorf("service %s does not exist", identifier)
		}

		res, err := template.RenderService(c.String("file"), existing)
		if err != nil {
			return err
		}
		return printResult(res)
	}

	bindings := map[string]interface{}{}
	if err := utils.YamlFile(c.String("configuration"), &bindings); err != nil {
		return err
	}

	res, err := template.RenderYaml(c.String("file"), bindings)
	if err != nil {
		return err
	}
	return printResult(res)
}

func (p *Plural) handleLuaTemplate(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	luaFile := c.String("lua-file")
	context := c.String("context")
	dir := c.String("dir")
	if dir == "" {
		dir = "."
	}

	if luaFile == "" {
		return fmt.Errorf("expected --lua-file flag")
	}

	luaStr, err := utils.ReadFile(luaFile)
	if err != nil {
		return err
	}

	ctx := map[string]interface{}{}
	if context != "" {
		if err := utils.YamlFile(context, &ctx); err != nil {
			return err
		}
	}

	values := map[interface{}]interface{}{}
	valuesFiles := []string{}

	dir, err = filepath.Abs(dir)
	if err != nil {
		return err
	}
	proc := luautils.NewProcessor(dir)
	defer proc.L.Close()

	// Register global values and valuesFiles in Lua
	valuesTable := proc.L.NewTable()
	proc.L.SetGlobal("values", valuesTable)

	valuesFilesTable := proc.L.NewTable()
	proc.L.SetGlobal("valuesFiles", valuesFilesTable)
	proc.L.SetGlobal("cluster", luautils.GoValueToLuaValue(proc.L, ctx["cluster"]))
	proc.L.SetGlobal("configuration", luautils.GoValueToLuaValue(proc.L, ctx["configuration"]))
	proc.L.SetGlobal("contexts", luautils.GoValueToLuaValue(proc.L, ctx["contexts"]))
	proc.L.SetGlobal("imports", luautils.GoValueToLuaValue(proc.L, ctx["imports"]))

	if err := proc.L.DoString(luaStr); err != nil {
		return err
	}

	if err := luautils.MapLua(proc.L.GetGlobal("values").(*lua.LTable), &values); err != nil {
		return err
	}

	if err := luautils.MapLua(proc.L.GetGlobal("valuesFiles").(*lua.LTable), &valuesFiles); err != nil {
		return err
	}

	result := map[string]interface{}{
		"values":      luautils.SanitizeValue(values),
		"valuesFiles": valuesFiles,
	}

	utils.Highlight("Final lua output:\n\n")
	utils.NewYAMLPrinter(result).PrettyPrint()
	return nil
}

func (p *Plural) handleCloneClusterService(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	cluster, err := p.ConsoleClient.GetCluster(common.GetIdAndName(c.Args().Get(0)))
	if err != nil {
		return err
	}
	if cluster == nil {
		return fmt.Errorf("could not find cluster %s", c.Args().Get(0))
	}

	serviceId, clusterName, serviceName, err := getServiceIdClusterNameServiceName(c.Args().Get(1))
	if err != nil {
		return err
	}

	attributes := gqlclient.ServiceCloneAttributes{
		Name:      c.String("name"),
		Namespace: lo.ToPtr(c.String("namespace")),
	}

	// TODO: DRY this up with service update
	var confArgs []string
	if c.IsSet("conf") {
		confArgs = append(confArgs, c.StringSlice("conf")...)
	}

	updateConfigurations := map[string]string{}
	for _, conf := range confArgs {
		configurationPair := strings.Split(conf, "=")
		if len(configurationPair) == 2 {
			updateConfigurations[configurationPair[0]] = configurationPair[1]
		}
	}
	for key, value := range updateConfigurations {
		attributes.Configuration = append(attributes.Configuration, &gqlclient.ConfigAttributes{
			Name:  key,
			Value: lo.ToPtr(value),
		})
	}

	sd, err := p.ConsoleClient.CloneService(cluster.ID, serviceId, serviceName, clusterName, attributes)
	if err != nil {
		return err
	}

	headers := []string{"Id", "Name", "Namespace", "Git Ref", "Git Folder", "Repo"}
	return utils.PrintTable([]*gqlclient.ServiceDeploymentFragment{sd}, headers, func(sd *gqlclient.ServiceDeploymentFragment) ([]string, error) {
		return []string{sd.ID, sd.Name, sd.Namespace, sd.Git.Ref, sd.Git.Folder, sd.Repository.URL}, nil
	})
}

func (p *Plural) handleUpdateClusterService(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	contextBindings := containers.NewSet[string]()
	serviceId, clusterName, serviceName, err := getServiceIdClusterNameServiceName(c.Args().Get(0))
	if err != nil {
		return err
	}

	existing, err := p.ConsoleClient.GetClusterService(serviceId, serviceName, clusterName)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("existing service deployment is empty")
	}
	existingConfigurations := map[string]string{}
	attributes := gqlclient.ServiceUpdateAttributes{
		Version: &existing.Version,
		Git: &gqlclient.GitRefAttributes{
			Ref:    existing.Git.Ref,
			Folder: existing.Git.Folder,
		},
		Configuration: []*gqlclient.ConfigAttributes{},
	}
	for _, context := range existing.Contexts {
		contextBindings.Add(context.ID)
	}

	if existing.DryRun != nil {
		attributes.DryRun = existing.DryRun
	}
	if existing.Kustomize != nil {
		attributes.Kustomize = &gqlclient.KustomizeAttributes{
			Path: existing.Kustomize.Path,
		}
	}

	for _, conf := range existing.Configuration {
		existingConfigurations[conf.Name] = conf.Value
	}

	v := c.String("version")
	if v != "" {
		attributes.Version = &v
	}
	if c.String("git-ref") != "" {
		attributes.Git.Ref = c.String("git-ref")
	}
	if c.String("git-folder") != "" {
		attributes.Git.Folder = c.String("git-folder")
	}
	var confArgs []string
	if c.IsSet("conf") {
		confArgs = append(confArgs, c.StringSlice("conf")...)
	}
	var contextArgs []string
	if c.IsSet("context-id") {
		contextArgs = append(contextArgs, c.StringSlice("context-id")...)
	}
	for _, context := range contextArgs {
		contextBindings.Add(context)
	}
	if contextBindings.Len() > 0 {
		attributes.ContextBindings = make([]*gqlclient.ContextBindingAttributes, 0)
		for _, context := range contextBindings.List() {
			attributes.ContextBindings = append(attributes.ContextBindings, &gqlclient.ContextBindingAttributes{
				ContextID: context,
			})
		}
	}

	updateConfigurations := map[string]string{}
	for _, conf := range confArgs {
		configurationPair := strings.Split(conf, "=")
		if len(configurationPair) == 2 {
			updateConfigurations[configurationPair[0]] = configurationPair[1]
		}
	}
	for k, v := range updateConfigurations {
		existingConfigurations[k] = v
	}
	for key, value := range existingConfigurations {
		attributes.Configuration = append(attributes.Configuration, &gqlclient.ConfigAttributes{
			Name:  key,
			Value: lo.ToPtr(value),
		})
	}
	if c.String("kustomize-folder") != "" {
		attributes.Kustomize = &gqlclient.KustomizeAttributes{
			Path: c.String("kustomize-folder"),
		}
	}
	if c.IsSet("dry-run") {
		dryRun := c.Bool("dry-run")
		attributes.DryRun = &dryRun
	}
	if c.IsSet("templated") {
		templated := c.Bool("templated")
		attributes.Templated = &templated
	}

	sd, err := p.ConsoleClient.UpdateClusterService(serviceId, serviceName, clusterName, attributes)
	if err != nil {
		return err
	}
	if sd == nil {
		return fmt.Errorf("returned object is nil")
	}

	headers := []string{"Id", "Name", "Namespace", "Git Ref", "Git Folder", "Repo"}
	return utils.PrintTable([]*gqlclient.ServiceDeploymentExtended{sd}, headers, func(sd *gqlclient.ServiceDeploymentExtended) ([]string, error) {
		return []string{sd.ID, sd.Name, sd.Namespace, sd.Git.Ref, sd.Git.Folder, sd.Repository.URL}, nil
	})
}

func (p *Plural) handleDescribeClusterService(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	serviceId, clusterName, serviceName, err := getServiceIdClusterNameServiceName(c.Args().Get(0))
	if err != nil {
		return err
	}
	existing, err := p.ConsoleClient.GetClusterService(serviceId, serviceName, clusterName)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("existing service deployment is empty")
	}
	output := c.String("o")
	switch {
	case output == "json":
		utils.NewJsonPrinter(existing).PrettyPrint()
		return nil
	case output == "yaml":
		utils.NewYAMLPrinter(existing).PrettyPrint()
		return nil
	case strings.HasPrefix(output, "jsonpath="):
		return utils.ParseJSONPath(output, existing)
	}

	desc, err := console.DescribeService(existing)
	if err != nil {
		return err
	}
	fmt.Print(desc)

	return nil
}

func (p *Plural) handleDeleteClusterService(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	serviceId, clusterName, serviceName, err := getServiceIdClusterNameServiceName(c.Args().Get(0))
	if err != nil {
		return err
	}

	svc, err := p.ConsoleClient.GetClusterService(serviceId, serviceName, clusterName)
	if err != nil {
		return err
	}
	if svc == nil {
		return fmt.Errorf("could not find service for %s", c.Args().Get(0))
	}

	deleted, err := p.ConsoleClient.DeleteClusterService(svc.ID)
	if err != nil {
		return fmt.Errorf("could not delete service: %w", err)
	}

	utils.Success("Service %s has been deleted successfully\n", deleted.DeleteServiceDeployment.Name)
	return nil
}

func (p *Plural) handleKickClusterService(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}
	serviceId, clusterName, serviceName, err := getServiceIdClusterNameServiceName(c.Args().Get(0))
	if err != nil {
		return err
	}
	svc, err := p.ConsoleClient.GetClusterService(serviceId, serviceName, clusterName)
	if err != nil {
		return err
	}
	if svc == nil {
		return fmt.Errorf("could not find service for %s", c.Args().Get(0))
	}
	kick, err := p.ConsoleClient.KickClusterService(serviceId, serviceName, clusterName)
	if err != nil {
		return err
	}
	utils.Success("Service %s has been sync successfully\n", kick.Name)
	return nil
}

func (p *Plural) handleTarballClusterService(c *cli.Context) error {
	serviceId, clusterName, serviceName, err := getServiceIdClusterNameServiceName(c.Args().Get(0))
	if err != nil {
		return fmt.Errorf("could not parse args: %w", err)
	}

	dir := c.String("dir")
	if err = utils.EnsureDir(dir); err != nil {
		return fmt.Errorf("could not ensure dir: %w", err)
	}

	if err = p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return fmt.Errorf("could not initialize console client: %w", err)
	}

	service, err := p.ConsoleClient.GetClusterService(serviceId, serviceName, clusterName)
	if err != nil {
		return fmt.Errorf("could not get service: %w", err)
	}
	if service == nil {
		return fmt.Errorf("could not get service for: %s", c.Args().Get(0))
	}
	if service.Tarball == nil {
		return fmt.Errorf("service %s does not have a tarball", service.Name)
	}

	deployToken, err := p.ConsoleClient.GetDeployToken(&service.Cluster.ID, nil)
	if err != nil {
		return fmt.Errorf("could not get deploy token: %w", err)
	}

	utils.Highlight("fetching tarball from %s\n", *service.Tarball)
	resp, err := utils.ReadRemoteFileWithRetries(*service.Tarball, deployToken, 3)
	if err != nil {
		return err
	}
	defer resp.Close()

	return fs.Untar(dir, resp)
}

type ServiceDeploymentAttributesConfiguration struct {
	Configuration []*gqlclient.ConfigAttributes
}

func getServiceIdClusterNameServiceName(input string) (serviceId, clusterName, serviceName *string, err error) {
	if strings.HasPrefix(input, "@") {
		i := strings.Trim(input, "@")
		split := strings.Split(i, "/")
		if len(split) != 2 {
			err = fmt.Errorf("expected format @{cluster-handle}/{serviceName}")
			return
		}
		clusterName = &split[0]
		serviceName = &split[1]
	} else {
		serviceId = &input
	}
	return
}

func validateFlag(ctx *cli.Context, name string, defaultVal string) (string, error) {
	res := ctx.String(name)
	if res == "" {
		if defaultVal == "" {
			return "", fmt.Errorf("expected --%s flag", name)
		}
		res = defaultVal
	}

	return res, nil
}
