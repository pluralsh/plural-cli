package scaffold

import (
	"os"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/output"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"gopkg.in/yaml.v2"
)

type Applications struct {
	Root string
}

func BuildApplications(root string) *Applications {
	return &Applications{Root: root}
}

func NewApplications() (*Applications, error) {
	root, err := git.Root()
	if err != nil {
		return nil, err
	}

	return BuildApplications(root), nil
}

func (apps *Applications) HelmValues(app string) (map[string]interface{}, error) {
	var res map[string]interface{}
	path := pathing.SanitizeFilepath(filepath.Join(apps.Root, app, "helm", app, "values.yaml"))
	content, err := os.ReadFile(path)
	if err != nil {
		return res, err
	}

	err = yaml.Unmarshal(content, &res)
	return res, err
}

func (apps *Applications) TerraformValues(app string) (map[string]interface{}, error) {
	out, err := output.Read(pathing.SanitizeFilepath(filepath.Join(apps.Root, app, "output.yaml")))
	return out.Terraform, err
}
