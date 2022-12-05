package template

import (
	"os"
	"path/filepath"

	"github.com/pluralsh/gqlclient"

	"github.com/imdario/mergo"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/pluralsh/plural/pkg/wkspace"
)

func BuildValuesFromTemplate(vals map[string]interface{}, prevVals map[string]map[string]interface{}, w *wkspace.Workspace) (map[string]map[string]interface{}, error) {
	globals := map[string]interface{}{}
	output := make(map[string]map[string]interface{})
	for _, chartInst := range w.Charts {
		chartName := chartInst.Chart.Name
		tplate := chartInst.Version.ValuesTemplate
		isLuaTemplate := chartInst.Version.TemplateType == gqlclient.TemplateTypeLua
		if w.Links != nil {
			if path, ok := w.Links.Helm[chartName]; ok {
				var err error
				tplate, err = utils.ReadFile(pathing.SanitizeFilepath(filepath.Join(path, "values.yaml.tpl")))
				if os.IsNotExist(err) {
					tplate, err = utils.ReadFile(pathing.SanitizeFilepath(filepath.Join(path, "values.yaml.lua")))
					if err != nil {
						return nil, err
					}
					isLuaTemplate = true
				}
			}
		}
		if isLuaTemplate {
			if err := FromLuaTemplate(vals, globals, output, chartName, tplate); err != nil {
				return nil, err
			}
		} else {
			if err := FromGoTemplate(vals, globals, output, chartName, tplate); err != nil {
				return nil, err
			}
		}
	}

	if err := mergo.Merge(&output, prevVals); err != nil {
		return nil, err
	}

	if len(globals) > 0 {
		output["global"] = globals
	}

	output["plrl"] = map[string]interface{}{
		"license": w.Installation.LicenseKey,
	}
	return output, nil
}
