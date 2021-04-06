package wkspace

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/michaeljguarino/forge/pkg/config"
	"github.com/michaeljguarino/forge/pkg/diff"
	"github.com/michaeljguarino/forge/pkg/manifest"
	"github.com/michaeljguarino/forge/pkg/provider"
	"github.com/michaeljguarino/forge/pkg/utils"
)

type MinimalWorkspace struct {
	Name     string
	Provider provider.Provider
	Config   *config.Config
	Manifest *manifest.ProjectManifest
}

func Minimal(name string) (*MinimalWorkspace, error) {
	root, err := utils.RepoRoot()
	if err != nil {
		return nil, err
	}

	path, _ := filepath.Abs(filepath.Join(root, name, "manifest.yaml"))
	var prov provider.Provider
	if utils.Exists(path) {
		manifest, err := manifest.Read(path)
		if err != nil {
			return nil, err
		}

		prov, err = provider.FromManifest(manifest)
		if err != nil {
			return nil, err
		}
	} else {
		prov, err = provider.Select(false)
		if err != nil {
			return nil, err
		}
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

func (m *MinimalWorkspace) BounceHelm() error {
	path, err := filepath.Abs(filepath.Join("helm", m.Name))
	if err != nil {
		return err
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

	namespace := m.Config.Namespace(m.Name)
	utils.Warn("helm diff upgrade --install --show-secrets --namespace %s %s %s\n", namespace, m.Name, path)
	return m.runDiff("helm", "diff", "upgrade", "--show-secrets", "--install", "--namespace", namespace, m.Name, path)
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
	root, err := utils.RepoRoot()
	if err != nil {
		return "", err
	}

	diffFolder, _ := filepath.Abs(filepath.Join(root, "diffs", m.Name))
	if err := os.MkdirAll(diffFolder, os.ModePerm); err != nil {
		return diffFolder, err
	}

	return diffFolder, err
}
