package wkspace

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/diff"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/output"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
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

	prov, err := provider.GetProvider()
	if err != nil {
		return nil, err
	}

	project, _ := manifest.ReadProject(pathing.SanitizeFilepath(filepath.Join(root, "workspace.yaml")))
	conf := config.Read()
	return &MinimalWorkspace{Name: name, Provider: prov, Config: &conf, Manifest: project}, nil
}

func FormatValues(w io.Writer, vals string, output *output.Output) (err error) {
	tmpl, err := template.New("gotpl").Parse(vals)
	if err != nil {
		return
	}
	err = tmpl.Execute(w, map[string]interface{}{"Import": *output})
	return
}

func templateVals(app, path string) (backup string, err error) {
	root, _ := utils.ProjectRoot()
	valsFile := pathing.SanitizeFilepath(filepath.Join(path, "values.yaml"))
	vals, err := utils.ReadFile(valsFile)
	if err != nil {
		return
	}

	out, err := output.Read(pathing.SanitizeFilepath(filepath.Join(root, app, "output.yaml")))
	if err != nil {
		out = output.New()
	}

	backup = fmt.Sprintf("%s.bak", valsFile)
	err = os.Rename(valsFile, backup)
	if err != nil {
		return
	}

	f, err := os.Create(valsFile)
	if err != nil {
		return
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	err = FormatValues(f, vals, out)
	return
}

func (m *MinimalWorkspace) BounceHelm(extraArgs ...string) error {
	path, err := filepath.Abs(pathing.SanitizeFilepath(filepath.Join("helm", m.Name)))
	if err != nil {
		return err
	}
	backup, err := templateVals(m.Name, path)
	if err == nil {
		defer func(oldpath, newpath string) {
			_ = os.Rename(oldpath, newpath)
		}(backup, pathing.SanitizeFilepath(filepath.Join(path, "values.yaml")))
	}

	namespace := m.Config.Namespace(m.Name)
	utils.Warn("helm upgrade --install --namespace %s %s %s %s\n", namespace, m.Name, path, strings.Join(extraArgs, " "))
	var args []string
	defaultArgs := []string{"upgrade", "--install", "--skip-crds", "--namespace", namespace, m.Name, path}
	args = append(args, defaultArgs...)
	args = append(args, extraArgs...)
	return utils.Cmd(m.Config,
		"helm", args...)
}

func (m *MinimalWorkspace) DiffHelm() error {
	path, err := filepath.Abs(m.Name)
	if err != nil {
		return err
	}
	backup, err := templateVals(m.Name, path)
	if err == nil {
		defer func(oldpath, newpath string) {
			_ = os.Rename(oldpath, newpath)
		}(backup, pathing.SanitizeFilepath(filepath.Join(path, "values.yaml")))
	}

	namespace := m.Config.Namespace(m.Name)
	utils.Warn("helm diff upgrade --install --show-secrets --reset-values --namespace %s %s %s\n", namespace, m.Name, path)
	if err := m.runDiff("helm", "diff", "upgrade", "--show-secrets", "--reset-values", "--install", "--namespace", namespace, m.Name, path); err != nil {
		utils.Note("helm diff failed, this command can be flaky, but let us know regardless")
	}
	return nil
}

func (m *MinimalWorkspace) DiffTerraform() error {
	return m.runDiff("terraform", "plan")
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
