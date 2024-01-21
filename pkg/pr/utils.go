package pr

import (
	"bytes"
	"os"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

func templateReplacement(data string, ctx map[string]interface{}) (string, error) {
	tpl, err := template.New("gotpl").Funcs(sprig.TxtFuncMap()).Parse(data)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, map[string]interface{}{"context": ctx}); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func replaceInPlace(path string, rep func(data []byte) ([]byte, error)) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	resData, err := rep(data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, resData, info.Mode())
}
