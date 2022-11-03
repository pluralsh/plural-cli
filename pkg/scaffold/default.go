package scaffold

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl"
	"github.com/pluralsh/plural/pkg/executor"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/pluralsh/plural/pkg/wkspace"
)

func Read(path string) (*Build, error) {
	fullpath := pathing.SanitizeFilepath(filepath.Join(path, "build.hcl"))
	contents, err := os.ReadFile(fullpath)
	build := Build{}
	if err != nil {
		return &build, err
	}

	err = hcl.Decode(&build, string(contents))
	if err != nil {
		return &build, err
	}

	return &build, nil
}

func Default(w *wkspace.Workspace, name string) (b *Build) {
	return &Build{
		Metadata: &Metadata{Name: name},
		Scaffolds: []*Scaffold{
			{
				Name: "terraform",
				Path: "terraform",
				Type: TF,
			},
			{
				Name: "crds",
				Type: CRD,
				Path: "crds",
			},
			{
				Name: "helm",
				Type: HELM,
				Path: pathing.SanitizeFilepath(filepath.Join("helm", name)),
				Preflight: []*executor.Step{
					{
						Name:    "update-deps",
						Command: "plural",
						Args:    []string{"wkspace", "helm-deps"},
						Target:  "Chart.yaml",
						Sha:     "",
					},
				},
			},
		},
	}
}
