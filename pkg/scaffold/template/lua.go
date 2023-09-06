package template

import (
	"fmt"

	"github.com/Masterminds/sprig/v3"
	"github.com/imdario/mergo"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"

	"github.com/pluralsh/plural/pkg/template"
	"github.com/pluralsh/plural/pkg/utils"
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
	outTable, ok := L.GetGlobal("output").(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("the output variable is missing in the lua script")
	}
	if err := utils.MapLua(outTable, &output); err != nil {
		return nil, err
	}

	return output, nil

}

func FromLuaTemplate(vals map[string]interface{}, globals map[string]interface{}, output map[string]map[string]interface{}, chartName, tplate string) error {
	var subVals = map[string]interface{}{}
	subVals, err := ExecuteLua(vals, tplate)
	if err != nil {
		return err
	}

	if _, exists := subVals["enabled"]; !exists {
		subVals["enabled"] = true
	}

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
