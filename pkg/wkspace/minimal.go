package wkspace

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	helmdiff "github.com/databus23/helm-diff/v3/diff"
	diffmanifest "github.com/databus23/helm-diff/v3/manifest"
	"github.com/google/go-cmp/cmp"
	"github.com/helm/helm-mapkubeapis/pkg/common"
	release "github.com/helm/helm-mapkubeapis/pkg/v3"
	"github.com/imdario/mergo"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	relutil "helm.sh/helm/v3/pkg/releaseutil"
	"helm.sh/helm/v3/pkg/storage/driver"
	"helm.sh/helm/v3/pkg/strvals"
	"sigs.k8s.io/yaml"

	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/diff"
	"github.com/pluralsh/plural/pkg/helm"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/output"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

const (
	valuesYaml           = "values.yaml"
	defaultValuesYaml    = "default-values.yaml"
	helm2TestSuccessHook = "test-success"
	helm3TestHook        = "test"
)

type MinimalWorkspace struct {
	Name       string
	Provider   provider.Provider
	Config     *config.Config
	Manifest   *manifest.ProjectManifest
	HelmConfig *action.Configuration
}

func Minimal(name string, helmConfig *action.Configuration) (*MinimalWorkspace, error) {
	root, err := git.Root()
	if err != nil {
		return nil, err
	}

	prov, err := provider.GetProvider()
	if err != nil {
		return nil, err
	}

	project, _ := manifest.ReadProject(pathing.SanitizeFilepath(filepath.Join(root, "workspace.yaml")))
	conf := config.Read()
	return &MinimalWorkspace{Name: name, Provider: prov, Config: &conf, Manifest: project, HelmConfig: helmConfig}, nil
}

func FormatValues(w io.Writer, vals string, output *output.Output) (err error) {
	tmpl, err := template.New("gotpl").Parse(vals)
	if err != nil {
		return
	}
	err = tmpl.Execute(w, map[string]interface{}{"Import": *output})
	return
}

func (m *MinimalWorkspace) BounceHelm(wait bool, skipArgs, setArgs, setJSONArgs []string) error {
	path, err := filepath.Abs(pathing.SanitizeFilepath(filepath.Join("helm", m.Name)))
	if err != nil {
		return err
	}
	defaultVals, err := getValues(m.Name)
	if err != nil {
		return err
	}

	for _, arg := range skipArgs {
		if err := strvals.ParseInto(arg, defaultVals); err != nil {
			return err
		}
	}
	for _, arg := range setArgs {
		if err := strvals.ParseInto(arg, defaultVals); err != nil {
			return err
		}
	}
	for _, arg := range setJSONArgs {
		if err := strvals.ParseJSON(arg, defaultVals); err != nil {
			return err
		}
	}

	namespace := m.Config.Namespace(m.Name)
	if m.HelmConfig == nil {
		m.HelmConfig, err = helm.GetActionConfig(namespace)
		if err != nil {
			return err
		}
	}

	utils.Warn("helm upgrade --install --skip-crds --namespace %s %s %s\n", namespace, m.Name, path)
	chart, err := loader.Load(path)
	if err != nil {
		return err
	}
	// If a release does not exist, install it.
	histClient := action.NewHistory(m.HelmConfig)
	histClient.Max = 5

	if _, err := histClient.Run(m.Name); errors.Is(err, driver.ErrReleaseNotFound) {
		instClient := action.NewInstall(m.HelmConfig)
		instClient.Namespace = namespace
		instClient.ReleaseName = m.Name
		instClient.SkipCRDs = true
		instClient.Timeout = time.Minute * 10
		instClient.Wait = wait

		if req := chart.Metadata.Dependencies; req != nil {
			if err := action.CheckDependencies(chart, req); err != nil {
				return err
			}
		}
		_, err := instClient.Run(chart, defaultVals)
		return err
	}

	client := action.NewUpgrade(m.HelmConfig)
	client.Namespace = namespace
	client.SkipCRDs = true
	client.Timeout = time.Minute * 10
	client.Wait = wait
	_, err = client.Run(m.Name, chart, defaultVals)
	if err != nil {
		current, errReleases := m.HelmConfig.Releases.Last(m.Name)
		if errReleases != nil {
			return errors.Wrap(err, fmt.Sprintf("can't get the last release %v", errReleases))
		}
		if !current.Info.Status.IsPending() {
			return err
		}
		deployedReleases, errDeployed := m.HelmConfig.Releases.ListDeployed()
		if errDeployed != nil {
			return errors.Wrap(err, fmt.Sprintf("can't get deployed releases %v", errDeployed))
		}
		rollback := action.NewRollback(m.HelmConfig)
		if len(deployedReleases) > 0 {
			relutil.Reverse(deployedReleases, relutil.SortByRevision)
			lastDeployed := deployedReleases[0].Version
			rollback.Version = lastDeployed
			utils.LogInfo().Printf("Rollback current: %d to last deployed %d \n", current.Version, deployedReleases[0].Version)
		}
		return rollback.Run(m.Name)
	}
	return err
}

func getValues(name string) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	defaultVals := make(map[string]interface{})

	path, err := getHelmPath(name)
	if err != nil {
		return nil, err
	}
	defaultValuesPath := pathing.SanitizeFilepath(filepath.Join(path, defaultValuesYaml))
	valuesPath := pathing.SanitizeFilepath(filepath.Join(path, valuesYaml))
	valsContent, err := os.ReadFile(valuesPath)
	if err != nil {
		return nil, err
	}
	valsContent, err = templateTerraformInputs(name, string(valsContent))
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(valsContent, &values); err != nil {
		return nil, err
	}
	if utils.Exists(defaultValuesPath) {
		defaultValsContent, err := os.ReadFile(defaultValuesPath)
		if err != nil {
			return nil, err
		}
		defaultValsContent, err = templateTerraformInputs(name, string(defaultValsContent))
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(defaultValsContent, &defaultVals); err != nil {
			return nil, err
		}
	}

	err = mergo.Merge(&defaultVals, values, mergo.WithOverride)
	if err != nil {
		return nil, err
	}
	return defaultVals, nil
}

func (m *MinimalWorkspace) TemplateHelm() error {
	path, err := filepath.Abs(pathing.SanitizeFilepath(filepath.Join("helm", m.Name)))
	if err != nil {
		return err
	}
	namespace := m.Config.Namespace(m.Name)
	manifest, err := m.getTemplate(false, false)
	if err != nil {
		return err
	}
	utils.Warn("helm template --skip-crds --namespace %s %s %s\n", namespace, m.Name, path)
	fmt.Printf("%s", manifest)
	return nil
}

func (m *MinimalWorkspace) DiffHelm() error {
	path, err := filepath.Abs(m.Name)
	if err != nil {
		return err
	}
	namespace := m.Config.Namespace(m.Name)
	utils.Warn("helm diff upgrade --install --show-secrets --reset-values  %s %s\n", m.Name, path)
	releaseManifest, err := m.getRelease()
	if err != nil {
		return err
	}
	installManifest, err := m.getTemplate(true, true)
	if err != nil {
		return err
	}

	diffFolder, err := m.constructDiffFolder()
	if err != nil {
		return err
	}
	outfile, err := os.Create(pathing.SanitizeFilepath(pathing.SanitizeFilepath(filepath.Join(diffFolder, "helm"))))
	if err != nil {
		return err
	}
	defer func(outfile *os.File) {
		_ = outfile.Close()
	}(outfile)

	mw := io.MultiWriter(os.Stdout, outfile)
	currentSpecs := diffmanifest.Parse(string(releaseManifest), namespace, false, helm3TestHook, helm2TestSuccessHook)
	newSpecs := diffmanifest.Parse(string(installManifest), namespace, false, helm3TestHook, helm2TestSuccessHook)
	helmdiff.Manifests(currentSpecs, newSpecs, &helmdiff.Options{
		OutputFormat:    "diff",
		OutputContext:   -1,
		StripTrailingCR: false,
		ShowSecrets:     true,
		SuppressedKinds: []string{},
		FindRenames:     0,
	}, mw)
	return nil
}

func (m *MinimalWorkspace) DiffTerraform() error {
	return m.runDiff("terraform", "plan")
}

func (m *MinimalWorkspace) MapKubeApis() error {
	namespace := m.Config.Namespace(m.Name)
	utils.Warn("helm mapkubeapis %s --namespace %s\n", m.Name, namespace)
	return mapKubeApis(m.Name, namespace)
}

func (m *MinimalWorkspace) runDiff(command string, args ...string) error {
	diffFolder, err := m.constructDiffFolder()
	if err != nil {
		return err
	}
	outfile, err := os.Create(pathing.SanitizeFilepath(pathing.SanitizeFilepath(filepath.Join(diffFolder, command))))
	if err != nil {
		return err
	}
	defer func(outfile *os.File) {
		_ = outfile.Close()
	}(outfile)

	cmd := exec.Command(command, args...)
	cmd.Stdout = &diff.TeeWriter{File: outfile}
	cmd.Stderr = os.Stdout
	return cmd.Run()
}

func (m *MinimalWorkspace) constructDiffFolder() (string, error) {
	root, err := git.Root()
	if err != nil {
		return "", err
	}

	diffFolder, _ := filepath.Abs(pathing.SanitizeFilepath(filepath.Join(root, "diffs", m.Name)))
	if err := os.MkdirAll(diffFolder, os.ModePerm); err != nil {
		return diffFolder, err
	}

	return diffFolder, err
}

func (m *MinimalWorkspace) getRelease() ([]byte, error) {
	namespace := m.Config.Namespace(m.Name)
	var err error
	if m.HelmConfig == nil {
		m.HelmConfig, err = helm.GetActionConfig(namespace)
		if err != nil {
			return nil, err
		}
	}
	client := action.NewGet(m.HelmConfig)
	rel, err := client.Run(m.Name)
	if err != nil {
		return nil, err
	}
	return []byte(rel.Manifest), nil
}

func (m *MinimalWorkspace) getTemplate(isUpgrade, validate bool) ([]byte, error) {
	path, err := getHelmPath(m.Name)
	if err != nil {
		return nil, err
	}
	defaultVals, err := getValues(m.Name)
	if err != nil {
		return nil, err
	}

	namespace := m.Config.Namespace(m.Name)

	if m.HelmConfig == nil {
		m.HelmConfig, err = helm.GetActionConfig(namespace)
		if err != nil {
			return nil, err
		}
	}

	return helm.Template(m.HelmConfig, m.Name, namespace, path, isUpgrade, validate, defaultVals)
}

func templateTerraformInputs(name, vals string) ([]byte, error) {
	root, _ := utils.ProjectRoot()
	out, err := output.Read(pathing.SanitizeFilepath(filepath.Join(root, name, "output.yaml")))
	if err != nil {
		out = output.New()
	}

	var buf bytes.Buffer
	buf.Grow(5 * 1024)

	err = FormatValues(&buf, vals, out)
	if err != nil {
		return nil, err
	}

	templatedData := buf.String()

	// This is a workaround for https://github.com/golang/go/issues/24963
	// In case terraform outputs are not there it will print '<no value>' and break helm templating
	sanitized := strings.ReplaceAll(templatedData, "<no value>", "")

	if len(templatedData) != len(sanitized) {
		msg := "Replaced '<no value>' with empty string to sanitize helm values:\n%s"
		utils.Warn(msg, cmp.Diff(templatedData, sanitized))
	}

	return []byte(sanitized), nil
}

func getHelmPath(name string) (string, error) {
	root, found := utils.ProjectRoot()
	if !found {
		return "", fmt.Errorf("couldn't find the root project path")
	}
	return filepath.Abs(pathing.SanitizeFilepath(filepath.Join(root, name, "helm", name)))
}

func mapKubeApis(name, namespace string) error {
	p, err := homedir.Expand("~/.plural")
	if err != nil {
		return err
	}
	mapFile := filepath.Join(p, "Map.yaml")
	if !utils.Exists(mapFile) {
		err := utils.DownloadFile(mapFile, "https://raw.githubusercontent.com/helm/helm-mapkubeapis/main/config/Map.yaml")
		if err != nil {
			return err
		}
	}
	options := common.MapOptions{
		DryRun:           false,
		KubeConfig:       common.KubeConfig{},
		MapFile:          mapFile,
		ReleaseName:      name,
		ReleaseNamespace: namespace,
	}
	return release.MapReleaseWithUnSupportedAPIs(options)
}
