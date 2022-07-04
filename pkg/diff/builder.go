package diff

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/pluralsh/plural/pkg/executor"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/rodaine/hclencoder"
)

type Diff struct {
	Metadata Metadata         `hcl:"metadata"`
	Steps    []*executor.Step `hcl:"step"`
}

type Metadata struct {
	Path string `hcl:"path"`
	Name string `hcl:"name"`
}

func GetDiff(path, name string) (*Diff, error) {
	fullpath := pathing.SanitizeFilepath(filepath.Join(path, name+".hcl"))
	contents, err := ioutil.ReadFile(fullpath)
	diff := Diff{}
	if err != nil {
		return &diff, nil
	}

	err = hcl.Decode(&diff, string(contents))
	return &diff, err
}

func (e *Diff) Execute() error {
	root, err := git.Root()
	if err != nil {
		return err
	}
	path := pathing.SanitizeFilepath(filepath.Join(root, "diffs"))
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}

	if err := utils.EmptyDirectory(path); err != nil {
		return err
	}

	ignore, err := e.IgnoreFile(root)
	if err != nil {
		return err
	}
	fmt.Printf("deploying %s, hold on to your butts\n", e.Metadata.Path)
	for i, step := range e.Steps {
		newSha, err := step.Execute(root, ignore)
		if err != nil {
			if err := e.Flush(root); err != nil {
				return err
			}

			return err
		}

		e.Steps[i].Sha = newSha
	}

	return e.Flush(root)
}

func (e *Diff) IgnoreFile(root string) ([]string, error) {
	ignorePath := pathing.SanitizeFilepath(filepath.Join(root, e.Metadata.Path, ".pluralignore"))
	contents, err := ioutil.ReadFile(ignorePath)
	if err != nil {
		return []string{}, err
	}

	ignore := strings.Split(string(contents), "\n")
	result := []string{}
	for _, prefix := range ignore {
		ignoreStr := strings.TrimSpace(prefix)
		if ignoreStr != "" {
			result = append(result, ignoreStr)
		}
	}

	return result, nil
}

func DefaultDiff(path string, prev *Diff) (e *Diff) {
	byName := make(map[string]*executor.Step)
	steps := []*executor.Step{
		{
			Name:    "terraform-init",
			Wkdir:   pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Command: "terraform",
			Args:    []string{"init"},
			Sha:     "",
		},
		{
			Name:    "terraform",
			Wkdir:   pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Command: "plural",
			Args:    []string{"wkspace", "terraform-diff", path},
			Sha:     "",
		},
		{
			Name:    "kube-init",
			Wkdir:   path,
			Target:  pluralfile(path, "NONCE"),
			Command: "plural",
			Args:    []string{"wkspace", "kube-init", path},
			Sha:     "",
		},
		{
			Name:    "helm",
			Wkdir:   pathing.SanitizeFilepath(filepath.Join(path, "helm")),
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "helm")),
			Command: "plural",
			Args:    []string{"wkspace", "helm-diff", path},
			Sha:     "",
		},
	}

	for _, step := range prev.Steps {
		byName[step.Name] = step
	}

	for _, step := range steps {
		prev, ok := byName[step.Name]
		if ok {
			step.Sha = prev.Sha
		}
		byName[step.Name] = step
	}

	// set up a topsort between the two orders of operations
	graph := utils.Graph(len(byName))
	for k := range byName {
		graph.AddNode(k)
	}

	for i := 0; i < len(steps)-1; i++ {
		graph.AddEdge(steps[i].Name, steps[i+1].Name)
	}

	for i := 0; i < len(prev.Steps)-1; i++ {
		graph.AddEdge(steps[i].Name, steps[i+1].Name)
	}

	finalizedSteps := []*executor.Step{}
	sorted, ok := graph.Topsort()
	if !ok {
		panic("deployfile cycle detected")
	}

	// dump the topsort to a list and use that from now on
	for _, name := range sorted {
		finalizedSteps = append(finalizedSteps, byName[name])
	}

	return &Diff{
		Metadata: Metadata{Path: path, Name: "diff"},
		Steps:    finalizedSteps,
	}
}

func (d *Diff) Flush(root string) error {
	io, err := hclencoder.Encode(&d)
	if err != nil {
		return err
	}

	path, _ := filepath.Abs(pathing.SanitizeFilepath(filepath.Join(root, d.Metadata.Path, d.Metadata.Name+".hcl")))
	return ioutil.WriteFile(path, io, 0644)
}

func pluralfile(base, name string) string {
	return pathing.SanitizeFilepath(filepath.Join(base, ".plural", name))
}
