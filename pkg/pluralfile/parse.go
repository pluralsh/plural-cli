package pluralfile

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Forgefile struct {
	Components []Component
	Repo       string
}

type ComponentName string

const (
	ARTIFACT    ComponentName = "artificat"
	TERRAFORM   ComponentName = "tf"
	HELM        ComponentName = "helm"
	RECIPE      ComponentName = "recipe"
	INTEGRATION ComponentName = "integration"
	CRD         ComponentName = "crd"
	IRD         ComponentName = "ird"
	COMMAND     ComponentName = "run"
)

type Component interface {
	Type() ComponentName
	Key() string
	Push(repo string, sha string) (string, error)
}

func (forge *Forgefile) Execute(f string, lock *Lockfile) (err error) {
	defer lock.Flush(f)
	for _, component := range forge.Components {
		key := component.Key()
		t := component.Type()
		sha := lock.getSha(t, key)
		newsha, err := component.Push(forge.Repo, sha)
		if err != nil {
			return err
		}
		lock.addSha(t, key, newsha)
	}

	return
}

func Parse(f string) (*Forgefile, error) {
	pluralfile, err := os.Open(f)
	forge := &Forgefile{}
	if err != nil {
		return forge, err
	}
	defer pluralfile.Close()

	scanner := bufio.NewScanner(pluralfile)
	for scanner.Scan() {
		line := scanner.Text()
		ignore, _ := regexp.MatchString(`^\s*$`, line)

		if ignore {
			continue
		}

		splitline := strings.Split(line, " ")

		switch strings.ToLower(splitline[0]) {
		case "repo":
			forge.Repo = splitline[1]
		case "helm":
			helms, err := expandGlob(splitline[1], func(targ string) Component {
				return &Helm{File: targ}
			})

			if err != nil {
				return forge, err
			}

			forge.Components = append(forge.Components, helms...)
		case "tf":
			tfs, err := expandGlob(splitline[1], func(targ string) Component {
				return &Terraform{File: targ}
			})

			if err != nil {
				return forge, err
			}

			forge.Components = append(forge.Components, tfs...)
		case "artifact":
			arts, err := expandGlob(splitline[1], func(targ string) Component {
				return &Artifact{File: targ, Platform: splitline[2], Arch: splitline[3]}
			})

			if err != nil {
				return forge, err
			}

			forge.Components = append(forge.Components, arts...)
		case "ird":
			irds, err := expandGlob(splitline[1], func(targ string) Component {
				return &ResourceDefinition{File: targ}
			})

			if err != nil {
				return forge, err
			}

			forge.Components = append(forge.Components, irds...)
		case "recipe":
			recipes, err := expandGlob(splitline[1], func(targ string) Component {
				return &Recipe{File: targ}
			})

			if err != nil {
				return forge, err
			}

			forge.Components = append(forge.Components, recipes...)
		case "integration":
			integs, err := expandGlob(splitline[1], func(targ string) Component {
				return &Integration{File: targ}
			})

			if err != nil {
				return forge, err
			}

			forge.Components = append(forge.Components, integs...)
		case "crd":
			chart := splitline[2]
			crds, err := expandGlob(splitline[1], func(targ string) Component {
				return &Crd{Chart: chart, File: targ}
			})

			if err != nil {
				return forge, err
			}
			forge.Components = append(forge.Components, crds...)
		case "run":
			cmd := splitline[1]
			args := splitline[2:]
			forge.Components = append(forge.Components, &Command{Command: cmd, Args: args})
		default:
			continue
		}
	}

	return forge, nil
}

func expandGlob(relpath string, toComponent func(path string) Component) ([]Component, error) {
	var comps []Component
	paths, err := filepath.Glob(relpath)
	if err != nil {
		return comps, err
	}

	for _, p := range paths {
		comps = append(comps, toComponent(p))
	}

	return comps, nil
}
