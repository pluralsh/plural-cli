package cd

import (
	"fmt"
	iofs "io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/console"

	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/polly/luautils"
	"github.com/samber/lo"
	"github.com/urfave/cli"
	lua "github.com/yuin/gopher-lua"
)

func (p *Plural) handleLuaTemplate(c *cli.Context) error {
	if err := p.InitConsoleClient(consoleToken, consoleURL); err != nil {
		return err
	}

	luaFile := c.String("lua-file")
	if luaFile == "" {
		return fmt.Errorf("expected --lua-file flag")
	}
	luaDir := c.String("lua-dir")

	context := c.String("context")
	serviceIdentifier := c.String("service")
	if !lo.IsEmpty(context) && !lo.IsEmpty(serviceIdentifier) {
		return fmt.Errorf("cannot specify both --context and --service flags")
	}

	dir := c.String("dir")
	if dir == "" {
		dir = "."
	}

	luaStr, err := utils.ReadFile(luaFile)
	if err != nil {
		return err
	}

	if luaDir != "" {
		luaFiles, err := luaFolder(luaDir)
		if err != nil {
			return err
		}

		luaStr = luaFiles + "\n\n" + luaStr
	}

	dir, err = filepath.Abs(dir)
	if err != nil {
		return err
	}

	bindings, err := luaBindings(p.ConsoleClient, context, serviceIdentifier)
	if err != nil {
		return err
	}

	result, err := executeLuaTemplate(luaStr, dir, bindings)
	if err != nil {
		return err
	}

	utils.Highlight("Final lua output:\n\n")
	utils.NewYAMLPrinter(result).PrettyPrint()
	return nil
}

// executeLuaTemplate runs the given Lua script with the provided bindings and working directory,
// and returns a map with "values" and "valuesFiles" keys.
func executeLuaTemplate(luaStr, dir string, bindings map[string]interface{}) (map[string]interface{}, error) {
	values := map[interface{}]interface{}{}
	valuesFiles := []string{}

	L := luautils.NewLuaState(dir)
	defer L.Close()

	// Register global values and valuesFiles in Lua
	valuesTable := L.NewTable()
	L.SetGlobal("values", valuesTable)

	valuesFilesTable := L.NewTable()
	L.SetGlobal("valuesFiles", valuesFilesTable)

	for name, binding := range bindings {
		L.SetGlobal(name, luautils.GoValueToLuaValue(L, binding))
	}

	if err := L.DoString(luaStr); err != nil {
		return nil, err
	}

	if err := luautils.MapLua(L.GetGlobal("values").(*lua.LTable), &values); err != nil {
		return nil, err
	}

	if err := luautils.MapLua(L.GetGlobal("valuesFiles").(*lua.LTable), &valuesFiles); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"values":      luautils.SanitizeValue(values),
		"valuesFiles": valuesFiles,
	}, nil
}

func luaFolder(folder string) (string, error) {
	luaFiles := make([]string, 0)
	if err := filepath.WalkDir(folder, func(path string, info iofs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(info.Name(), ".lua") {
			luaPath, err := filepath.Rel(folder, path)
			if err != nil {
				return err
			}
			luaFiles = append(luaFiles, luaPath)
		}

		return nil
	}); err != nil {
		return "", fmt.Errorf("failed to walk lua folder %s: %w", folder, err)
	}

	sort.Slice(luaFiles, func(i, j int) bool {
		return luaFiles[i] < luaFiles[j]
	})

	luaFileContents := make([]string, 0)
	for _, file := range luaFiles {
		luaContents, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("failed to read lua file %s: %w", file, err)
		}
		luaFileContents = append(luaFileContents, string(luaContents))
	}

	return strings.Join(luaFileContents, "\n\n"), nil
}

func luaBindings(client console.ConsoleClient, contextPath, serviceIdentifier string) (context map[string]interface{}, err error) {
	if serviceIdentifier != "" {
		service, err := getService(client, serviceIdentifier)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"configuration": luaConfigurationBinding(service),
			"cluster":       luaClusterBinding(service.Cluster),
			"contexts":      luaContextsBinding(service),
			"imports":       luaImportsBinding(service),
			"service":       luaServiceBinding(service),
		}, nil
	}

	if contextPath != "" {
		err = utils.YamlFile(contextPath, &context)
		return
	}

	return
}

func luaConfigurationBinding(svc *client.ServiceDeploymentExtended) map[string]string {
	res := map[string]string{}
	for _, config := range svc.Configuration {
		res[config.Name] = config.Value
	}
	return res
}

func luaClusterBinding(cluster *client.BaseClusterFragment) map[string]interface{} {
	res := map[string]interface{}{
		"ID":             cluster.ID,
		"Self":           cluster.Self,
		"Handle":         cluster.Handle,
		"Name":           cluster.Name,
		"Version":        cluster.Version,
		"CurrentVersion": cluster.CurrentVersion,
		"KasUrl":         cluster.KasURL,
		"Tags":           luaClusterTagsBinding(cluster.Tags),
		"Metadata":       cluster.Metadata,
		"Distro":         cluster.Distro,
		//"ConsoleDNS":     args.ConsoleDNS(),
	}
	for k, v := range res {
		res[strings.ToLower(k)] = v
	}
	res["kasUrl"] = cluster.KasURL
	res["currentVersion"] = cluster.CurrentVersion
	return res
}

func luaClusterTagsBinding(tags []*client.ClusterTags) map[string]string {
	res := map[string]string{}
	for _, tag := range tags {
		res[tag.Name] = tag.Value
	}
	return res
}

func luaContextsBinding(svc *client.ServiceDeploymentExtended) map[string]map[string]interface{} {
	res := map[string]map[string]interface{}{}
	for _, context := range svc.Contexts {
		res[context.Name] = context.Configuration
	}
	return res
}

func luaImportsBinding(svc *client.ServiceDeploymentExtended) map[string]map[string]string {
	res := map[string]map[string]string{}
	for _, imp := range svc.Imports {
		res[imp.Stack.Name] = map[string]string{}
		for _, out := range imp.Outputs {
			res[imp.Stack.Name][out.Name] = out.Value
		}
	}
	return res
}

func luaServiceBinding(svc *client.ServiceDeploymentExtended) map[string]interface{} {
	res := map[string]interface{}{
		"Name":      svc.Name,
		"Namespace": svc.Namespace,
	}
	for k, v := range res {
		res[strings.ToLower(k)] = v
	}
	//if svc.Helm != nil {
	//	helm := map[string]interface{}{
	//		"Values":              svc.Helm.Values,
	//		"ValuesFiles":         svc.Helm.ValuesFiles,
	//		"LuaScript":           svc.Helm.LuaScript,
	//		"LuaFile":             svc.Helm.LuaFile,
	//		"LuaFolder":           svc.Helm.LuaFolder,
	//		"KustomizePostrender": svc.Helm.KustomizePostrender,
	//	}
	//
	//	for k, f := range helm {
	//		helm[strings.ToLower(k)] = f
	//	}
	//	res["helm"] = helm
	//	res["Helm"] = helm
	//}
	return res
}
