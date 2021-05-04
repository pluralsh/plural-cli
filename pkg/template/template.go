package template

import (
	"bytes"
	"io"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/pluralsh/plural/pkg/utils"
)

func MakeTemplate(tmplate string) (*template.Template, error) {
	funcs := sprig.TxtFuncMap()
	funcs["genAESKey"] = utils.GenAESKey
	funcs["repoRoot"] = repoRoot
	funcs["repoName"] = repoName
	funcs["repoUrl"] = repoUrl
	funcs["branchName"] = branchName
	funcs["createWebhook"] = createWebhook
	funcs["dumpConfig"] = dumpConfig
	funcs["dumpAesKey"] = dumpAesKey
	funcs["readLine"] = readLine
	funcs["readLineDefault"] = readLineDefault
	funcs["readFile"] = readFile
	funcs["homeDir"] = homeDir
	funcs["knownHosts"] = knownHosts
	funcs["dedupe"] = dedupe
	funcs["probe"] = probe
	funcs["importValue"] = importValue
	return template.New("gotpl").Funcs(funcs).Parse(tmplate)
}

func RenderTemplate(wr io.Writer, tmplate string, ctx map[string]interface{}) error {
	tmpl, err := MakeTemplate(tmplate)
	if err != nil {
		return err
	}
	return tmpl.Execute(wr, map[string]interface{}{"Values": ctx})
}

func RenderString(tmplate string, ctx map[string]interface{}) (string, error) {
	var buffer bytes.Buffer
	err := RenderTemplate(&buffer, tmplate, ctx)
	return buffer.String(), err
}
