package wkspace

import (
	"io"
	"text/template"

	"github.com/pluralsh/plural-cli/pkg/output"
)

func FormatValues(w io.Writer, vals string, output *output.Output) (err error) {
	tmpl, err := template.New("gotpl").Parse(vals)
	if err != nil {
		return
	}
	err = tmpl.Execute(w, map[string]interface{}{"Import": *output})
	return
}
