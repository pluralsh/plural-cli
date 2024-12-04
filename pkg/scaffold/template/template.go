package template

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pluralsh/gqlclient"
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/output"
	"github.com/pluralsh/plural-cli/pkg/template"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/pathing"
	"github.com/pluralsh/plural-cli/pkg/wkspace"
	"gopkg.in/yaml.v2"
)

func templateInfo(path string) (t gqlclient.TemplateType, contents string, err error) {
	gopath := pathing.SanitizeFilepath(filepath.Join(path, "values.yaml.tpl"))
	if utils.Exists(gopath) {
		contents, err = utils.ReadFile(gopath)
		t = gqlclient.TemplateTypeGotemplate
		return
	}

	luapath := pathing.SanitizeFilepath(filepath.Join(path, "values.yaml.lua"))
	if utils.Exists(gopath) {
		contents, err = utils.ReadFile(luapath)
		t = gqlclient.TemplateTypeLua
		return
	}

	err = fmt.Errorf("could not find values.yaml.tpl or values.yaml.lua in directory, perhaps your link is to the wrong folder?")
	return
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
		"Values":        map[string]interface{}{},
		"Configuration": map[string]map[string]interface{}{},
		"License":       "example-license",
		"Region":        "region",
		"Project":       "example",
		"Cluster":       "cluster",
		"Provider":      "provider",
		"Config":        conf,
		"Context":       map[string]interface{}{},
	}
}
