package template

import (
	"github.com/Masterminds/sprig/v3"
	"github.com/imdario/mergo"
	"github.com/pluralsh/plural/pkg/template"
	"github.com/pluralsh/plural/pkg/utils"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

func ExecuteLua(vals map[string]interface{}, tplate string) (map[string]interface{}, error) {
	output := map[string]interface{}{}
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
		return nil, err
	}
	if err := utils.MapLua(L.GetGlobal("output").(*lua.LTable), &output); err != nil {
		return nil, err
	}

	return output, nil

}

func FromLuaTemplate(vals map[string]interface{}, globals map[string]interface{}, output map[string]map[string]interface{}, chartName, tplate string) error {
	subVals, err := ExecuteLua(vals, tplate)
	if err != nil {
		return err
	}
	subVals["enabled"] = true

	// need to handle globals in a dedicated way
	if glob, ok := subVals["global"]; ok {
		globMap := utils.CleanUpInterfaceMap(glob.(map[interface{}]interface{}))
		if err := mergo.Merge(&globals, globMap); err != nil {
			return err
		}
		delete(subVals, "global")
	}

	output[chartName] = subVals

	return nil
}
