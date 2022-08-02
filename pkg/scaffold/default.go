package scaffold

import (
	"path/filepath"

	hcl "github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/pluralsh/plural/pkg/executor"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/pluralsh/plural/pkg/wkspace"
)

func Read(path string) (*Build, error) {
	fullpath := pathing.SanitizeFilepath(filepath.Join(path, "build.hcl"))
	build := Build{}

	err := hcl.DecodeFile(fullpath, nil, &build)
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
