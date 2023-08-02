package template

import (
	"bytes"

	"github.com/imdario/mergo"
	"github.com/pluralsh/plural/pkg/template"
	"github.com/pluralsh/plural/pkg/utils"
	"gopkg.in/yaml.v2"
)

func FromGoTemplate(vals map[string]interface{}, globals map[string]interface{}, output map[string]map[string]interface{}, chartName, tplate string) error {
	var buf bytes.Buffer
	buf.Grow(5 * 1024)

	tmpl, err := template.MakeTemplate(tplate)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(&buf, vals); err != nil {
		return err
	}

	var subVals = map[string]interface{}{}
	if err := yaml.Unmarshal(buf.Bytes(), &subVals); err != nil {
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
	buf.Reset()
	return nil
}
