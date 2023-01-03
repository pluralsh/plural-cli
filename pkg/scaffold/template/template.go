package template

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pluralsh/gqlclient"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/output"
	"github.com/pluralsh/plural/pkg/template"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/pluralsh/plural/pkg/wkspace"
	"gopkg.in/yaml.v2"
)

func BuildValuesFromTemplate(vals map[string]interface{}, w *wkspace.Workspace) (map[string]map[string]interface{}, error) {
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

	if len(globals) > 0 {
		output["global"] = globals
	}

	output["plrl"] = map[string]interface{}{
		"license": w.Installation.LicenseKey,
	}
	return output, nil
}

func TmpValuesFile(path string) (f *os.File, err error) {
	conf := config.Read()
	if strings.HasSuffix(path, "lua") {
		return luaTmpValuesFile(path, &conf)
	}

	return goTmpValuesFile(path, &conf)

}

func luaTmpValuesFile(path string, conf *config.Config) (f *os.File, err error) {
	valuesTmpl, err := utils.ReadFile(path)
	if err != nil {
		return
	}
	f, err = os.CreateTemp("", "values.yaml")
	if err != nil {
		return
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	vals := genDefaultValues(conf)

	output, err := ExecuteLua(vals, valuesTmpl)
	if err != nil {
		return nil, err
	}

	io, err := yaml.Marshal(output)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(io))
	_, err = f.Write(io)
	if err != nil {
		return nil, err
	}
	return
}

func goTmpValuesFile(path string, conf *config.Config) (f *os.File, err error) {
	valuesTmpl, err := utils.ReadFile(path)
	if err != nil {
		return
	}
	tmpl, err := template.MakeTemplate(valuesTmpl)
	if err != nil {
		return
	}

	vals := genDefaultValues(conf)
	var buf bytes.Buffer

	if err = tmpl.Execute(&buf, vals); err != nil {
		return
	}

	f, err = os.CreateTemp("", "values.yaml")
	if err != nil {
		return
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	fmt.Println(buf.String())
	err = wkspace.FormatValues(f, buf.String(), output.New())
	return
}

func genDefaultValues(conf *config.Config) map[string]interface{} {
	return map[string]interface{}{
		"Values":   map[string]interface{}{},
		"License":  "example-license",
		"Region":   "region",
		"Project":  "example",
		"Cluster":  "cluster",
		"Provider": "provider",
		"Config":   conf,
		"Context":  map[string]interface{}{},
	}
}
