package scaffold

import (
	"path/filepath"

	"github.com/Masterminds/sprig/v3"
	"github.com/imdario/mergo"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/template"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/pluralsh/plural/pkg/wkspace"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

func FromLuaTemplate(vals map[string]interface{}, globals map[string]interface{}, values map[string]map[string]interface{}, w *wkspace.Workspace, chartInst *api.ChartInstallation) error {
	tplate := chartInst.Version.ValuesTemplate
	if w.Links != nil {
		if path, ok := w.Links.Helm[chartInst.Chart.Name]; ok {
			var err error
			tplate, err = utils.ReadFile(pathing.SanitizeFilepath(filepath.Join(path, "values.yaml.lua")))
			if err != nil {
				return err
			}
		}
	}
	subVals := map[string]interface{}{}
	subVals["enabled"] = true

	L := lua.NewState()
	defer L.Close()

	L.SetGlobal("Var", luar.New(L, vals))

	for name, function := range template.GetFuncMap() {
		L.SetGlobal(name, luar.New(L, function))
	}
	for name, function := range sprig.GenericFuncMap() {
		L.SetGlobal(name, luar.New(L, function))
	}

	if err := L.DoString(tplate); err != nil {
		return err
	}
	if err := utils.MapLua(L.GetGlobal("valuesYaml").(*lua.LTable), &subVals); err != nil {
		return err
	}
	// need to handle globals in a dedicated way
	if glob, ok := subVals["global"]; ok {
		globMap := utils.CleanUpInterfaceMap(glob.(map[interface{}]interface{}))
		if err := mergo.Merge(&globals, globMap); err != nil {
			return err
		}
		delete(subVals, "global")
	}

	values[chartInst.Chart.Name] = subVals

	return nil
}
