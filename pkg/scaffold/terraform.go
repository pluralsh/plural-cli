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

const outputTemplate = `output "{{ .Name }}" {
	value = module.{{ .Module }}.{{ .Value }}
	sensitive = true
}
`

func (scaffold *Scaffold) handleTerraform(wk *wkspace.Workspace) error {
	repo := wk.Installation.Repository
	providerCtx := buildContext(wk, repo.Name, wk.Terraform)
	backend, err := wk.Provider.CreateBackend(repo.Name, providerCtx)
	if err != nil {
		return err
	}

	apps, err := NewApplications()
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
	ctx, _ := wk.Context.Repo(repo.Name)
	links := wk.Links
	for i, tfInst := range wk.Terraform {
		tf := tfInst.Terraform
		linkPath := ""
		if links != nil {
			if path, ok := links.Terraform[tf.Name]; ok {
				linkPath = path
			}
		}

		var buf bytes.Buffer
		buf.Grow(5 * 1024)
		plate := tfInst.Version.ValuesTemplate
		if linkPath != "" {
			var err error
			plate, err = utils.ReadFile(filepath.Join(linkPath, "terraform.tfvars"))
			if err != nil {
				return err
			}
		}

		tmpl, err := template.MakeTemplate(plate)
		if err != nil {
			return err
		}
		values := map[string]interface{}{
			"Values":        ctx,
			"Configuration": wk.Context.Configuration,
			"Cluster":       wk.Provider.Cluster(),
			"Project":       wk.Provider.Project(),
			"Namespace":     wk.Config.Namespace(repo.Name),
			"Region":        wk.Provider.Region(),
			"Context":       wk.Provider.Context(),
			"Applications":  apps,
		}
		if err := tmpl.Execute(&buf, values); err != nil {
			return err
		}

		module := make(map[string]interface{})
		module["name"] = tf.Name
		if linkPath != "" {
			module["path"] = linkPath
		} else {
			module["path"] = "./" + tf.Name
		}

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

	if err := scaffold.buildOutputs(wk); err != nil {
		return err
	}

	secrets := buildTfSecrets(wk.Terraform)
	if err := buildSecrets(filepath.Join(scaffold.Root, ".gitattributes"), secrets); err != nil {
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

func (scaffold *Scaffold) buildOutputs(wk *wkspace.Workspace) error {
	var buf bytes.Buffer
	buf.Grow(5 * 1024)

	tmp, err := template.MakeTemplate(outputTemplate)
	if err != nil {
		return err
	}

	for _, tfInst := range wk.Terraform {
		tfName := tfInst.Terraform.Name
		for name, value := range tfInst.Version.Dependencies.Outputs {
			err = tmp.Execute(&buf, map[string]interface{}{"Name": name, "Value": value, "Module": tfName})
			if err != nil {
				return err
			}
			buf.WriteString("\n\n")
		}
	}

	outputFile := filepath.Join(scaffold.Root, "outputs.tf")
	return utils.WriteFile(outputFile, buf.Bytes())
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

func buildTfSecrets(installations []*api.TerraformInstallation) []string {
	res := []string{}
	for _, inst := range installations {
		res = append(res, inst.Version.Dependencies.Secrets...)
	}
	return res
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
