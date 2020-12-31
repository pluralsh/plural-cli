package scaffold

import (
	"path/filepath"
	"io/ioutil"
	"github.com/hashicorp/hcl"
	"github.com/michaeljguarino/forge/executor"
)

func Read(path string) (*Build, error) {
	fullpath := filepath.Join(path, "build.hcl")
	contents, err := ioutil.ReadFile(fullpath)
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

func Default(name string) (b *Build) {
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
				Path: filepath.Join("helm", name),
				Preflight: []*executor.Step{
					{
						Name:    "add-repo",
						Command: "helm",
						Args:    []string{"repo", "add", name, repoUrl(name)},
						Target:  "requirements.yaml",
						Sha:     "",
					},
					{
						Name:    "update-deps",
						Command: "helm",
						Args:    []string{"dependency", "update"},
						Target:  "requirements.yaml",
						Sha:     "",
					},
				},
			},
		},
	}
}