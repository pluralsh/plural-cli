package scaffold

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	ttpl "text/template"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/provider"
	scftmpl "github.com/pluralsh/plural/pkg/scaffold/template"
	"github.com/pluralsh/plural/pkg/template"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/errors"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/pluralsh/plural/pkg/wkspace"
	"gopkg.in/yaml.v2"
)

type dependency struct {
	Name       string
	Version    string
	Repository string
	Condition  string
}

type chart struct {
	ApiVersion   string `yaml:"apiVersion"`
	Name         string
	Description  string
	Version      string
	AppVersion   string `yaml:"appVersion"`
	Dependencies []dependency
}

func (s *Scaffold) handleHelm(wk *wkspace.Workspace) error {
	if err := s.createChart(wk); err != nil {
		return err
	}

	if err := s.buildChartValues(wk); err != nil {
		return err
	}

	return nil
}

func (s *Scaffold) chartDependencies(w *wkspace.Workspace) []dependency {
	dependencies := make([]dependency, len(w.Charts))
	repo := w.Installation.Repository
	for i, chartInstallation := range w.Charts {
		dependencies[i] = dependency{
			chartInstallation.Chart.Name,
			chartInstallation.Version.Version,
			repoUrl(w, repo.Name, chartInstallation.Chart.Name),
			fmt.Sprintf("%s.enabled", chartInstallation.Chart.Name),
		}
	}
	sort.SliceStable(dependencies, func(i, j int) bool {
		return dependencies[i].Name < dependencies[j].Name
	})
	return dependencies
}

func Notes(installation *api.Installation) error {
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	if installation.Repository != nil && installation.Repository.Notes == "" {
		return nil
	}

	context, err := manifest.ReadContext(manifest.ContextPath())
	if err != nil {
		return err
	}

	prov, err := provider.GetProvider()
	if err != nil {
		return err
	}

	repo := installation.Repository.Name
	ctx, _ := context.Repo(installation.Repository.Name)

	vals := map[string]interface{}{
		"Values":        ctx,
		"Configuration": context.Configuration,
		"License":       installation.LicenseKey,
		"OIDC":          installation.OIDCProvider,
		"Region":        prov.Region(),
		"Project":       prov.Project(),
		"Cluster":       prov.Cluster(),
		"Config":        config.Read(),
		"Provider":      prov.Name(),
		"Context":       prov.Context(),
		"Applications":  BuildApplications(repoRoot),
	}

	if context.Globals != nil {
		vals["Globals"] = context.Globals
	}

	if context.SMTP != nil {
		vals["SMTP"] = context.SMTP.Configuration()
	}

	if installation.AcmeKeyId != "" {
		vals["Acme"] = map[string]string{
			"KeyId":  installation.AcmeKeyId,
			"Secret": installation.AcmeSecret,
		}
	}

	apps := &Applications{Root: repoRoot}
	values, err := apps.HelmValues(repo)
	if err != nil {
		return err
	}

	for k, v := range values {
		vals[k] = v
	}

	tmpl, err := template.MakeTemplate(installation.Repository.Notes)
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
	valuesFile := pathing.SanitizeFilepath(filepath.Join(s.Root, "values.yaml"))
	defaultValuesFile := pathing.SanitizeFilepath(filepath.Join(s.Root, "default-values.yaml"))
	defaultPrevVals, _ := prevValues(defaultValuesFile)
	prevVals, _ := prevValues(valuesFile)

	if !utils.Exists(valuesFile) {
		if err := os.WriteFile(valuesFile, []byte(""), 0644); err != nil {
			return err
		}
	}

	conf := config.Read()

	apps, err := NewApplications()
	if err != nil {
		return err
	}

	proj, err := manifest.FetchProject()
	if err != nil {
		return err
	}

	vals := map[string]interface{}{
		"Values":        ctx,
		"Configuration": w.Context.Configuration,
		"License":       w.Installation.LicenseKey,
		"OIDC":          w.Installation.OIDCProvider,
		"Region":        w.Provider.Region(),
		"Project":       w.Provider.Project(),
		"Cluster":       w.Provider.Cluster(),
		"Config":        conf,
		"Provider":      w.Provider.Name(),
		"Context":       w.Provider.Context(),
		"Network":       proj.Network,
		"Applications":  apps,
	}

	if w.Context.SMTP != nil {
		vals["SMTP"] = w.Context.SMTP.Configuration()
	}

	if w.Context.Globals != nil {
		vals["Globals"] = w.Context.Globals
	}

	if w.Installation.AcmeKeyId != "" {
		vals["Acme"] = map[string]string{
			"KeyId":  w.Installation.AcmeKeyId,
			"Secret": w.Installation.AcmeSecret,
		}
	}

	// get previous values from default-values.yaml if exists otherwise from values.yaml
	if utils.Exists(defaultValuesFile) {
		for k, v := range defaultPrevVals {
			vals[k] = v
		}
	} else {
		for k, v := range prevVals {
			vals[k] = v
		}
	}

	defaultValues, err := scftmpl.BuildValuesFromTemplate(vals, w)
	if err != nil {
		return err
	}

	io, err := yaml.Marshal(defaultValues)
	if err != nil {
		return err
	}

	mapValues, err := getValues(valuesFile)
	if err != nil {
		return err
	}
	patchValues, err := utils.PatchInterfaceMap(defaultValues, mapValues)
	if err != nil {
		return err
	}

	values, err := yaml.Marshal(patchValues)
	if err != nil {
		return err
	}
	if len(patchValues) == 0 {
		values = []byte{}
	}
	if err := utils.WriteFile(valuesFile, values); err != nil {
		return err
	}

	return utils.WriteFile(defaultValuesFile, io)
}

func getValues(path string) (map[string]map[string]interface{}, error) {
	values := map[string]map[string]interface{}{}
	valuesFromFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(valuesFromFile, &values); err != nil {
		return nil, err
	}
	return values, nil
}

func prevValues(filename string) (map[string]map[string]interface{}, error) {
	vals := make(map[string]map[interface{}]interface{})
	parsed := make(map[string]map[string]interface{})
	if !utils.Exists(filename) {
		return parsed, nil
	}

	contents, err := os.ReadFile(filename)
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

func (s *Scaffold) createChart(w *wkspace.Workspace) error {
	repo := w.Installation.Repository
	if len(w.Charts) == 0 {
		return utils.HighlightError(fmt.Errorf("No charts installed for this repository. You might need to run `plural bundle install %s <bundle-name>`.", repo.Name))
	}

	version := "0.1.0"
	filename := pathing.SanitizeFilepath(filepath.Join(s.Root, ChartfileName))

	if utils.Exists(filename) {
		content, err := os.ReadFile(filename)
		if err != nil {
			return errors.ErrorWrap(err, "Failed to read existing Chart.yaml")
		}

		chart := chart{}
		if err := yaml.Unmarshal(content, &chart); err != nil {
			return errors.ErrorWrap(err, "Existing Chart.yaml has invalid yaml formatting")
		}

		version = chart.Version
	}

	appVersion := appVersion(w.Charts)
	chart := &chart{
		ApiVersion:   "v2",
		Name:         repo.Name,
		Description:  fmt.Sprintf("A helm chart for %s", repo.Name),
		Version:      version,
		AppVersion:   appVersion,
		Dependencies: s.chartDependencies(w),
	}

	chartFile, err := yaml.Marshal(chart)
	if err != nil {
		return err
	}

	if err := utils.WriteFile(filename, chartFile); err != nil {
		return err
	}

	files := []struct {
		path    string
		content []byte
		force   bool
	}{
		{
			// .helmignore
			path:    pathing.SanitizeFilepath(filepath.Join(s.Root, IgnorefileName)),
			content: []byte(defaultIgnore),
		},
		{
			// NOTES.txt
			path:    pathing.SanitizeFilepath(filepath.Join(s.Root, NotesName)),
			content: []byte(defaultNotes),
			force:   true,
		},
		{
			// templates/secret.yaml
			path:    pathing.SanitizeFilepath(filepath.Join(s.Root, LicenseSecretName)),
			content: []byte(licenseSecret),
			force:   true,
		},
		{
			// templates/licnse.yaml
			path:    pathing.SanitizeFilepath(filepath.Join(s.Root, LicenseCrdName)),
			content: []byte(fmt.Sprintf(license, repo.Name)),
			force:   true,
		},
	}

	for _, file := range files {
		if !file.force {
			if _, err := os.Stat(file.path); err == nil {
				// File exists and is okay. Skip it.
				continue
			}
		}
		if err := utils.WriteFile(file.path, file.content); err != nil {
			return err
		}
	}

	// remove old requirements.yaml files to fully migrate to helm v3
	reqsFile := pathing.SanitizeFilepath(filepath.Join(s.Root, "requirements.yaml"))
	if utils.Exists(reqsFile) {
		if err := os.Remove(reqsFile); err != nil {
			return err
		}
	}

	tpl, err := ttpl.New("gotpl").Parse(defaultApplication)
	if err != nil {
		return err
	}

	var appBuffer bytes.Buffer
	vars := map[string]string{
		"Name":        repo.Name,
		"Version":     appVersion,
		"Description": repo.Description,
		"Icon":        repo.Icon,
		"DarkIcon":    repo.DarkIcon,
	}
	if err := tpl.Execute(&appBuffer, vars); err != nil {
		return err
	}
	appBuffer.WriteString(appTemplate)

	if err := utils.WriteFile(pathing.SanitizeFilepath(filepath.Join(s.Root, ApplicationName)), appBuffer.Bytes()); err != nil {
		return err
	}

	// Need to add the ChartsDir explicitly as it does not contain any file OOTB
	if err := os.MkdirAll(pathing.SanitizeFilepath(filepath.Join(s.Root, ChartsDir)), 0755); err != nil {
		return err
	}

	return nil
}

func repoUrl(w *wkspace.Workspace, repo string, chart string) string {
	if w.Links != nil {
		if path, ok := w.Links.Helm[chart]; ok {
			return fmt.Sprintf("file://%s", path)
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
