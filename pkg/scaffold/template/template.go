package template

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/output"
	"github.com/pluralsh/plural-cli/pkg/template"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/wkspace"
	"gopkg.in/yaml.v2"
)

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
