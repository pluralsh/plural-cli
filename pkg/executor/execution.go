package executor

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/rodaine/hclencoder"
)

type Execution struct {
	Metadata Metadata `hcl:"metadata"`
	Steps    []*Step  `hcl:"step"`
}

type Metadata struct {
	Path string `hcl:"path"`
	Name string `hcl:"name"`
}

const (
	pluralIgnore = `terraform/.terraform`
)

func Ignore(root string) error {
	ignoreFile := pathing.SanitizeFilepath(filepath.Join(root, ".pluralignore"))
	return ioutil.WriteFile(ignoreFile, []byte(pluralIgnore), 0644)
}

func GetExecution(path, name string) (*Execution, error) {
	fullpath := pathing.SanitizeFilepath(filepath.Join(path, name+".hcl"))
	contents, err := ioutil.ReadFile(fullpath)
	ex := Execution{}
	if err != nil {
		return &ex, err
	}

	err = hcl.Decode(&ex, string(contents))
	if err != nil {
		return &ex, err
	}

	return &ex, nil
}

func (e *Execution) Execute(verbose bool) error {
	root, err := git.Root()
	if err != nil {
		return err
	}

	ignore, err := e.IgnoreFile(root)
	if err != nil {
		return err
	}

	fmt.Printf("deploying %s.  This may take a while, so hold on to your butts\n", e.Metadata.Path)
	for i, step := range e.Steps {
		prev := step.Verbose
		if verbose {
			step.Verbose = true
		}

		newSha, err := step.Execute(root, ignore)
		step.Verbose = prev
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

func (e *Execution) IgnoreFile(root string) ([]string, error) {
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

func DefaultExecution(path string, prev *Execution) (e *Execution) {
	byName := make(map[string]*Step)
	steps := defaultSteps(path)

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
		graph.AddEdge(prev.Steps[i].Name, prev.Steps[i+1].Name)
	}

	finalizedSteps := []*Step{}
	sorted, ok := graph.Topsort()
	if !ok {
		panic("deployfile cycle detected")
	}

	// dump the topsort to a list and use that from now on
	for _, name := range sorted {
		finalizedSteps = append(finalizedSteps, byName[name])
	}

	return &Execution{
		Metadata: Metadata{Path: path, Name: "deploy"},
		Steps:    finalizedSteps,
	}
}

func (e *Execution) Flush(root string) error {
	io, err := hclencoder.Encode(&e)
	if err != nil {
		return err
	}

	path, _ := filepath.Abs(pathing.SanitizeFilepath(filepath.Join(root, e.Metadata.Path, e.Metadata.Name+".hcl")))
	return ioutil.WriteFile(path, io, 0644)
}

func pluralfile(base, name string) string {
	return pathing.SanitizeFilepath(filepath.Join(base, ".plural", name))
}
