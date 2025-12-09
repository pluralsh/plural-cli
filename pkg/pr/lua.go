package pr

import (
	"fmt"
	iofs "io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"dario.cat/mergo"
	"github.com/pluralsh/polly/luautils"
	lua "github.com/yuin/gopher-lua"
	"k8s.io/apimachinery/pkg/util/json"
)

func executeLua(spec *PrTemplateSpec, ctx map[string]interface{}) error {
	if spec.Lua == nil {
		return nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	L := luautils.NewLuaState(dir)
	defer L.Close()

	// Register global values and valuesFiles in Lua
	prAutomation := L.NewTable()
	L.SetGlobal("prAutomation", prAutomation)
	L.SetGlobal("context", luautils.GoValueToLuaValue(L, ctx))

	parent := dir
	if spec.Lua.ExternalDir != "" {
		parent = spec.Lua.ExternalDir
	}

	var luaString string
	if len(spec.Lua.Script) > 0 {
		luaString = spec.Lua.Script
	}

	if spec.Lua.Folder != "" && len(spec.Lua.Folder) > 0 {
		luaFolder, err := luaFolder(parent, spec.Lua.Folder)
		if err != nil {
			return err
		}
		luaString = luaFolder + "\n\n" + luaString
	}

	if luaString == "" {
		return fmt.Errorf("no lua script folder provided")
	}

	// Execute the Lua script
	if err := L.DoString(luaString); err != nil {
		return err
	}

	prSpec := luautils.ToGoValue(L.GetGlobal("prAutomation").(*lua.LTable))
	mapPrSpec, ok := luautils.SanitizeValue(prSpec).(map[string]interface{})
	if !ok {
		return fmt.Errorf("failed to parse prAutomation table")
	}

	additionalCtx := map[string]interface{}{}
	if err := luautils.MapLua(L.GetGlobal("context").(*lua.LTable), &additionalCtx); err != nil {
		return err
	}

	if err := mergo.Merge(&ctx, additionalCtx, mergo.WithAppendSlice, mergo.WithOverride); err != nil {
		return err
	}

	return merge(spec, mapPrSpec)
}

func merge(spec *PrTemplateSpec, new map[string]interface{}) error {
	jsonStr, err := json.Marshal(new)
	if err != nil {
		return err
	}

	newSpec := PrTemplateSpec{}
	if err := json.Unmarshal(jsonStr, &newSpec); err != nil {
		return err
	}

	return mergo.Merge(spec, newSpec, mergo.WithAppendSlice, mergo.WithOverride)
}

func luaFolder(parent, folder string) (string, error) {
	luaFiles := make([]string, 0)

	if err := filepath.WalkDir(filepath.Join(parent, folder), func(path string, info iofs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(info.Name(), ".lua") {
			luaPath, err := filepath.Rel(parent, path)
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
		luaContents, err := os.ReadFile(filepath.Join(parent, file))
		if err != nil {
			return "", fmt.Errorf("failed to read lua file %s: %w", file, err)
		}
		luaFileContents = append(luaFileContents, string(luaContents))
	}

	return strings.Join(luaFileContents, "\n\n"), nil
}
