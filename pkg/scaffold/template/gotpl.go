package template

import (
	"bytes"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v2"

	"github.com/pluralsh/plural/pkg/template"
	"github.com/pluralsh/plural/pkg/utils"
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
	subVals["enabled"] = true
	if err := yaml.Unmarshal(buf.Bytes(), &subVals); err != nil {
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

	output[chartName] = subVals
	buf.Reset()
	return nil
}
