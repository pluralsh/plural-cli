package stacks

import (
	_ "embed"
	"path"
	"strings"
	"text/template"

	"github.com/pluralsh/plural-cli/pkg/utils"
)

//go:embed templates/_override.tf.gotmpl
var overrideTemplateText string

type OverrideTemplateInput struct {
	Address       string
	LockAddress   string
	UnlockAddress string
	Actor         string
	DeployToken   string
}

func GenerateOverrideTemplate(input *OverrideTemplateInput, dir string) (fileName string, err error) {
	fileName = "_override.tf"
	tmpl, err := template.New(fileName).Parse(overrideTemplateText)
	if err != nil {
		return "", err
	}

	out := new(strings.Builder)
	err = tmpl.Execute(out, input)
	if err != nil {
		return "", err
	}

	return fileName, utils.WriteFile(path.Join(dir, fileName), []byte(out.String()))
}
