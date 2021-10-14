package scaffold

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	ttpl "text/template"
	"strings"

	"github.com/imdario/mergo"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/template"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/wkspace"
	"github.com/pluralsh/plural/pkg/manifest"
	"gopkg.in/yaml.v2"
)

type dependency struct {
	Name       string
	Version    string
	Repository string
}

type chart struct {
	ApiVersion  string `yaml:"apiVersion"`
	Name 			  string
	Description string
	Version 		string
	AppVersion  string `yaml:"appVersion"`
	Dependencies []dependency
}

func (s *Scaffold) handleHelm(wk *wkspace.Workspace) error {
	repo := wk.Installation.Repository

	err := s.createChart(wk, repo.Name)
	if err != nil {
		return err
	}

	if err := s.buildChartValues(wk); err != nil {
		return err
	}

	return nil
}

func (s *Scaffold) createChartDependencies(w *wkspace.Workspace, name string) error {
	dependencies := s.chartDependencies(w, name)
	io, err := yaml.Marshal(map[string][]dependency{"dependencies": dependencies})
	if err != nil {
		return err
	}

	requirementsFile := filepath.Join(s.Root, "requirements.yaml")
	return utils.WriteFile(requirementsFile, io)
}

func (s *Scaffold) chartDependencies(w *wkspace.Workspace, name string) []dependency {
	dependencies := make([]dependency, len(w.Charts))
	repo := w.Installation.Repository
	for i, chartInstallation := range w.Charts {
		dependencies[i] = dependency{
			chartInstallation.Chart.Name,
			chartInstallation.Version.Version,
			repoUrl(w, repo.Name),
		}
	}
	return dependencies
}

func Notes(w *wkspace.Workspace) error {
	repoRoot, err := utils.RepoRoot()
	if err != nil {
		return err
	}

	if w.Installation.Repository.Notes == "" {
		return nil
	}

	repo := w.Installation.Repository.Name
	ctx, _ := w.Context.Repo(w.Installation.Repository.Name)
	valuesFile := filepath.Join(repoRoot, repo, "helm", repo, "values.yaml")
	prevVals, _ := prevValues(valuesFile)
	conf := config.Read()
	vals := map[string]interface{}{
		"Values":        ctx,
		"Configuration": w.Context.Configuration,
		"License":       w.Installation.License,
		"OIDC":          w.Installation.OIDCProvider,
		"Region":        w.Provider.Region(),
		"Project":       w.Provider.Project(),
		"Cluster":       w.Provider.Cluster(),
		"Config":        conf,
		"Provider":      w.Provider.Name(),
		"Context":       w.Provider.Context(),
	}

	if (w.Context.SMTP != nil) {
		vals["SMTP"] = w.Context.SMTP.Configuration()
	}

	if (w.Installation.AcmeKeyId != "") {
		vals["Acme"] = map[string]string{
			"KeyId": w.Installation.AcmeKeyId,
			"Secret": w.Installation.AcmeSecret,
		}
	}

	for k, v := range prevVals {
		vals[k] = v
	}

	tmpl, err := template.MakeTemplate(w.Installation.Repository.Notes)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.Grow(5 * 1024)
	if err := tmpl.Execute(&buf, vals); err != nil {
		return err
	}

	fmt.Println(buf.String())
	return nil
}

func (s *Scaffold) buildChartValues(w *wkspace.Workspace) error {
	ctx, _ := w.Context.Repo(w.Installation.Repository.Name)
	var buf bytes.Buffer
	values := make(map[string]map[string]interface{})
	buf.Grow(5 * 1024)

	valuesFile := filepath.Join(s.Root, "values.yaml")
	prevVals, _ := prevValues(valuesFile)
	conf := config.Read()
	globals := map[string]interface{}{}

	proj, err := manifest.FetchProject()
	if err != nil {
		return err
	}

	for _, chartInst := range w.Charts {
		plate := chartInst.Version.ValuesTemplate
		if w.Links != nil {
			if path, ok := w.Links.Helm[chartInst.Chart.Name]; ok {
				var err error
				plate, err = utils.ReadFile(filepath.Join(path, "values.yaml.tpl"))
				if err != nil {
					return err
				}
			}
		}

		tmpl, err := template.MakeTemplate(plate)
		if err != nil {
			return err
		}

		vals := map[string]interface{}{
			"Values":        ctx,
			"Configuration": w.Context.Configuration,
			"License":       w.Installation.License,
			"OIDC":          w.Installation.OIDCProvider,
			"Region":        w.Provider.Region(),
			"Project":       w.Provider.Project(),
			"Cluster":       w.Provider.Cluster(),
			"Config":        conf,
			"Provider":      w.Provider.Name(),
			"Context":       w.Provider.Context(),
			"Network":       proj.Network,
		}

		if (w.Context.SMTP != nil) {
			vals["SMTP"] = w.Context.SMTP.Configuration()
		}

		if (w.Installation.AcmeKeyId != "") {
			vals["Acme"] = map[string]string{
				"KeyId": w.Installation.AcmeKeyId,
				"Secret": w.Installation.AcmeSecret,
			}
		}

		for k, v := range prevVals {
			vals[k] = v
		}

		if err := tmpl.Execute(&buf, vals); err != nil {
			return err
		}

		var subVals map[string]interface{}
		if err := yaml.Unmarshal(buf.Bytes(), &subVals); err != nil {
			return err
		}

		// need to handle globals in a dedicated way
		if glob, ok := subVals["global"]; ok {
			globMap := utils.CleanUpInterfaceMap(glob.(map[interface{}]interface{}))
			mergo.Merge(&globals, globMap)
			delete(subVals, "global")
		}

		values[chartInst.Chart.Name] = subVals
		buf.Reset()
	}

	if err := mergo.Merge(&values, prevVals); err != nil {
		return err
	}

	if len(globals) > 0 {
		values["global"] = globals
	}

	io, err := yaml.Marshal(values)
	if err != nil {
		fmt.Println("Invalid yaml:\n")
		fmt.Println(values)
		return err
	}

	return utils.WriteFile(valuesFile, io)
}

func prevValues(filename string) (map[string]map[string]interface{}, error) {
	vals := make(map[string]map[interface{}]interface{})
	parsed := make(map[string]map[string]interface{})
	if !utils.Exists(filename) {
		return parsed, nil
	}

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return parsed, err
	}
	if err := yaml.Unmarshal(contents, &vals); err != nil {
		return parsed, err
	}

	for k, v := range vals {
		parsed[k] = utils.CleanUpInterfaceMap(v)
	}

	return parsed, nil
}

func (s *Scaffold) createChart(w *wkspace.Workspace, name string) error {
	repo := w.Installation.Repository
	if len(w.Charts) == 0 {
		return utils.HighlightError(fmt.Errorf("No charts installed for this repository, you might need to run `plural bundle install %s <bundle-name>`", repo.Name))
	}

	appVersion := appVersion(w.Charts)
	chart := &chart{
		ApiVersion: "v2",
		Name: repo.Name,
		Description: fmt.Sprintf("A helm chart for %s", repo.Name),
		Version: "0.1.0",
		AppVersion: appVersion,
		Dependencies: s.chartDependencies(w, name),
	}

	chartFile, err := yaml.Marshal(chart)
	if err != nil {
		return err
	}

	if err := utils.WriteFile(filepath.Join(s.Root, ChartfileName), chartFile); err != nil {
		return err
	}

	files := []struct {
		path    string
		content []byte
	}{
		{
			// .helmignore
			path:    filepath.Join(s.Root, IgnorefileName),
			content: []byte(defaultIgnore),
		},
		{
			// NOTES.txt
			path:    filepath.Join(s.Root, NotesName),
			content: []byte(defaultNotes),
		},
	}

	for _, file := range files {
		if _, err := os.Stat(file.path); err == nil {
			// File exists and is okay. Skip it.
			continue
		}
		if err := utils.WriteFile(file.path, file.content); err != nil {
			return err
		}
	}

	// remove old requirements.yaml files to fully migrate to helm v3
	reqsFile := filepath.Join(s.Root, "requirements.yaml")
	if utils.Exists(reqsFile) {
		os.Remove(reqsFile)
	}

	tpl, err := ttpl.New("gotpl").Parse(defaultApplication)
	if err != nil {
		return err
	}

	var appBuffer bytes.Buffer
	vars := map[string]string{
		"Name": repo.Name,
		"Version": appVersion,
		"Description": repo.Description,
		"Icon": repo.Icon,
		"DarkIcon": repo.DarkIcon,
	}
	if err := tpl.Execute(&appBuffer, vars); err != nil {
		return err
	}


	if err := utils.WriteFile(filepath.Join(s.Root, ApplicationName), appBuffer.Bytes()); err != nil {
		return err
	}

	// Need to add the ChartsDir explicitly as it does not contain any file OOTB
	if err := os.MkdirAll(filepath.Join(s.Root, ChartsDir), 0755); err != nil {
		return err
	}

	return nil
}

func repoUrl(w *wkspace.Workspace, repo string) string {
	if w.Links != nil {
		if path, ok := w.Links.Helm[repo]; ok {
			return fmt.Sprintf("file:%s", path)
		}
	}
	url := strings.ReplaceAll(w.Config.BaseUrl(), "https", "cm")
	return fmt.Sprintf("%s/cm/%s", url, repo)
}

func appVersion(charts []*api.ChartInstallation) string {
	for _, inst := range charts {
		if inst.Chart.Dependencies.Application {
			return inst.Version.Version
		}
	}

	return "0.1.0"
}
