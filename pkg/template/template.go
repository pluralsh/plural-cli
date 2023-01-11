package template

import (
	"bytes"
	"io"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/pluralsh/plural/pkg/utils"
)

func GetFuncMap() template.FuncMap {
	funcs := sprig.TxtFuncMap()
	funcs["genAESKey"] = utils.GenAESKey
	funcs["repoRoot"] = repoRoot
	funcs["repoName"] = repoName
	funcs["repoUrl"] = repoUrl
	funcs["branchName"] = branchName
	funcs["dumpConfig"] = dumpConfig
	funcs["dumpAesKey"] = dumpAesKey
	funcs["readLine"] = readLine
	funcs["readPassword"] = readPassword
	funcs["readLineDefault"] = readLineDefault
	funcs["readFile"] = readFile
	funcs["homeDir"] = homeDir
	funcs["knownHosts"] = knownHosts
	funcs["dedupe"] = dedupe
	funcs["dedupeObj"] = dedupeObj
	funcs["secret"] = secret
	funcs["probe"] = probe
	funcs["importValue"] = importValue
	funcs["namespace"] = namespace
	funcs["toYaml"] = toYaml
	funcs["fileExists"] = fileExists
	funcs["pathJoin"] = pathJoin
	funcs["eabCredential"] = eabCredential
	funcs["encrypt"] = encrypt
	return funcs
}

func MakeTemplate(tmplate string) (*template.Template, error) {
	return template.New("gotpl").Funcs(GetFuncMap()).Parse(tmplate)
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
