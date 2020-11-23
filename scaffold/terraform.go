package scaffold

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/michaeljguarino/forge/api"
	"github.com/michaeljguarino/forge/template"
	"github.com/michaeljguarino/forge/utils"
	"github.com/michaeljguarino/forge/wkspace"
)

const moduleTemplate = `module "{{ .Values.name }}" {
  source = "{{ .Values.path }}"

### BEGIN MANUAL SECTION <<{{ .Values.name }}>>
{{ .Values.Manual }}
### END MANUAL SECTION <<{{ .Values.name }}>>

{{ .Values.conf | nindent 2 }}
{{ range $key, $val := .Values.deps }}
  {{ $key }} = module.{{ $val }}
{{- end }}
}
`

func (scaffold *Scaffold) handleTerraform(wk *wkspace.Workspace) error {
	repo := wk.Installation.Repository
	ctx := wk.Installation.Context
	var modules = make([]string, len(wk.Terraform)+1)
	backend, err := wk.Provider.CreateBackend(repo.Name)
	if err != nil {
		return err
	}

	if err := scaffold.untarModules(wk); err != nil {
		return err
	}
	mainFile := filepath.Join(scaffold.Root, "main.tf")
	contents, err := utils.ReadFile(mainFile)
	if err != nil {
		contents = ""
	}

	modules[0] = backend
	for i, tfInst := range wk.Terraform {
		tf := tfInst.Terraform

		var buf bytes.Buffer
		buf.Grow(5 * 1024)
		tmpl, err := template.MakeTemplate(tf.ValuesTemplate)
		if err != nil {
			return err
		}
		if err := tmpl.Execute(
			&buf, map[string]interface{}{"Values": ctx, "Cluster": wk.Provider.Cluster()}); err != nil {
			return err
		}

		module := make(map[string]interface{})
		module["name"] = tf.Name
		module["path"] = "./" + tf.Name
		module["conf"] = buf.String()
		module["deps"] = tf.Dependencies.Wirings.Terraform
		module["Manual"] = manualSection(contents, tf.Name)

		var moduleBuf bytes.Buffer
		moduleBuf.Grow(1024)
		if err := template.RenderTemplate(&moduleBuf, moduleTemplate, module); err != nil {
			return err
		}

		modules[i+1] = moduleBuf.String()

		valuesFile := filepath.Join(scaffold.Root, tf.Name, "terraform.tfvars")
		os.Remove(valuesFile)

		moduleBuf.Reset()
		buf.Reset()
	}

	if err := utils.WriteFile(mainFile, []byte(strings.Join(modules, "\n\n"))); err != nil {
		return err
	}

	return nil
}

// TODO: move to some sort of scaffold util?
func (scaffold *Scaffold) untarModules(wk *wkspace.Workspace) error {
	utils.Highlight("unpacking %d module(s)", len(wk.Terraform))
	for _, tfInst := range wk.Terraform {
		tf := tfInst.Terraform
		v := tfInst.Version
		path := filepath.Join(scaffold.Root, tf.Name)
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			fmt.Print("\n")
			return err
		}

		if err := untar(&v, &tf, path); err != nil {
			fmt.Print("\n")
			return err
		}
		fmt.Print(".")
	}

	utils.Success("\u2713\n")
	return nil
}

func untar(v *api.Version, tf *api.Terraform, dir string) error {
	resp, err := http.Get(v.Package)
	if err != nil {
		return err
	}

	return utils.Untar(resp.Body, dir, tf.Name)
}

func manualSection(contents, name string) string {
	re := regexp.MustCompile(fmt.Sprintf(`(?s)### BEGIN MANUAL SECTION <<%s>>(.*)### END MANUAL SECTION <<%s>>`, name, name))
	matches := re.FindStringSubmatch(contents)
	if len(matches) > 0 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}
