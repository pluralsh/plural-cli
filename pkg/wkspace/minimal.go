package wkspace

import (
	"os"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/diff"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/output"
)

type MinimalWorkspace struct {
	Name     string
	Provider provider.Provider
	Config   *config.Config
	Manifest *manifest.ProjectManifest
}

func Minimal(name string) (*MinimalWorkspace, error) {
	root, err := git.Root()
	if err != nil {
		return nil, err
	}

	path, _ := filepath.Abs(filepath.Join(root, name, "manifest.yaml"))
	prov, err := provider.Bootstrap(path, false)
	if err != nil {
		return nil, err
	}

	project, _ := manifest.ReadProject(filepath.Join(root, "workspace.yaml"))
	conf := config.Read()
	return &MinimalWorkspace{Name: name, Provider: prov, Config: &conf, Manifest: project}, nil
}

func (m *MinimalWorkspace) HelmInit(clientOnly bool) error {
	home, _ := os.UserHomeDir()
	helmRepos := filepath.Join(home, ".helm", "repository", "repositories.yaml")
	if !utils.Exists(helmRepos) && clientOnly {
		return utils.Cmd(m.Config, "helm", "init", "--client-only")
	}
	if !clientOnly && !utils.InKubernetes() {
		return utils.Cmd(m.Config, "helm", "init", "--wait", "--service-account=tiller")
	}

	return nil
}

const pullSecretName = "forgecreds"
const repoName = "dkr.piazza.app"

func (m *MinimalWorkspace) EnsurePullCredentials() error {
	name := m.Name
	if err := utils.Cmd(m.Config, "kubectl", "get", "secret", pullSecretName, "--namespace", name); err != nil {
		token := m.Config.Token
		username := m.Config.Email

		return utils.Cmd(m.Config,
			"kubectl", "create", "secret", "docker-registry", pullSecretName,
			"--namespace", name, "--docker-username", username, "--docker-password", token, "--docker-server", repoName)
	}

	return nil
}

func FormatValues(w io.Writer, vals string, output *output.Output) (err error) {
	tmpl, err := template.New("gotpl").Parse(vals)
	if err != nil { return }
	err = tmpl.Execute(w, map[string]interface{}{"Import": *output})
	return
}

func templateVals(app, path string) (backup string, err error) {
	root, _ := utils.ProjectRoot()
	valsFile := filepath.Join(path, "values.yaml")
	vals, err := utils.ReadFile(valsFile)
	if err != nil { return }

	out, err := output.Read(filepath.Join(root, app, "output.yaml"))
	if err != nil { 
		out = output.New()
	}

	backup = fmt.Sprintf("%s.bak", valsFile)
	err = os.Rename(valsFile, backup)
	if err != nil { return }

	f, err := os.Create(valsFile)
	if err != nil { return }
	defer f.Close()

	err = FormatValues(f, vals, out)
	return
}

func (m *MinimalWorkspace) BounceHelm() error {
	path, err := filepath.Abs(filepath.Join("helm", m.Name))
	if err != nil {
		return err
	}
	backup, err := templateVals(m.Name, path)
	if err == nil {
		defer os.Rename(backup, filepath.Join(path, "values.yaml"))
	}

	namespace := m.Config.Namespace(m.Name)
	utils.Warn("helm upgrade --install --namespace %s %s %s\n", namespace, m.Name, path)
	return utils.Cmd(m.Config,
		"helm", "upgrade", "--install", "--skip-crds", "--namespace", namespace, m.Name, path)
}

func (m *MinimalWorkspace) DiffHelm() error {
	path, err := filepath.Abs(m.Name)
	if err != nil {
		return err
	}
	backup, err := templateVals(m.Name, path)
	if err == nil {
		defer os.Rename(backup, filepath.Join(path, "values.yaml"))
	}

	namespace := m.Config.Namespace(m.Name)
	utils.Warn("helm diff upgrade --install --show-secrets --reset-values --namespace %s %s %s\n", namespace, m.Name, path)
	return m.runDiff("helm", "diff", "upgrade", "--show-secrets", "--reset-values", "--install", "--namespace", namespace, m.Name, path)
}

func (m *MinimalWorkspace) DiffTerraform() error {
	return m.runDiff("terraform", "plan")
}

func (m *MinimalWorkspace) runDiff(command string, args ...string) error {
	diffFolder, err := m.constructDiffFolder()
	outfile, err := os.Create(filepath.Join(diffFolder, command))
	if err != nil {
		return err
	}
	defer outfile.Close()

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

	diffFolder, _ := filepath.Abs(filepath.Join(root, "diffs", m.Name))
	if err := os.MkdirAll(diffFolder, os.ModePerm); err != nil {
		return diffFolder, err
	}

	return diffFolder, err
}
