package scaffold

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/template"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/wkspace"
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
	providerCtx := buildContext(wk, repo.Name, wk.Terraform)
	backend, err := wk.Provider.CreateBackend(repo.Name, providerCtx)
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

	var modules = make([]string, len(wk.Terraform)+1)
	modules[0] = backend
	ctx := wk.Installation.Context
	for i, tfInst := range wk.Terraform {
		tf := tfInst.Terraform

		var buf bytes.Buffer
		buf.Grow(5 * 1024)
		tmpl, err := template.MakeTemplate(tf.ValuesTemplate)
		if err != nil {
			return err
		}
		values := map[string]interface{}{
			"Values": ctx, 
			"Cluster": wk.Provider.Cluster(),
			"Project": wk.Provider.Project(),
			"Namespace": wk.Config.Namespace(repo.Name),
		}
		if err := tmpl.Execute(&buf, values); err != nil {
			return err
		}

		module := make(map[string]interface{})
		module["name"] = tf.Name
		module["path"] = "./" + tf.Name
		module["conf"] = buf.String()
		if tf.Dependencies != nil && tf.Dependencies.Wirings != nil {
			module["deps"] = tf.Dependencies.Wirings.Terraform
		} else {
			module["deps"] = map[string]interface{}{}
		}
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
	length := len(wk.Terraform)
	utils.Highlight("unpacking %d %s", len(wk.Terraform), utils.Pluralize("module", "modules", length))
	for _, tfInst := range wk.Terraform {
		tf := tfInst.Terraform
		v := tfInst.Version
		path := filepath.Join(scaffold.Root, tf.Name)
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			fmt.Print("\n")
			return err
		}

		if err := untar(v, tf, path); err != nil {
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

func buildContext(wk *wkspace.Workspace, repo string, installations []*api.TerraformInstallation) map[string]interface{} {
	ctx := map[string]interface{}{
		"Namespace": wk.Config.Namespace(repo),
	}

	for _, inst := range installations {
		for k, v := range inst.Version.Dependencies.ProviderWirings {
			if k == "cluster" {
				ctx["Cluster"] = v
			}
			ctx[k] = v
		}
	}

	return ctx
}
