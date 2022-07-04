package pluralfile

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/shlex"
)

type Pluralfile struct {
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
	TAG         ComponentName = "tag"
	REPO_ATTRS  ComponentName = "attrs"
)

type Component interface {
	Type() ComponentName
	Key() string
	Push(repo string, sha string) (string, error)
}

func (plrl *Pluralfile) Execute(_ string, lock *Lockfile) (err error) {
	defer func(plrl *Pluralfile, lock *Lockfile) {
		_ = plrl.Flush(lock)
	}(plrl, lock)
	for _, component := range plrl.Components {
		key := component.Key()
		t := component.Type()
		sha := lock.getSha(t, key)
		newsha, err := component.Push(plrl.Repo, sha)
		if err != nil {
			return err
		}
		lock.addSha(t, key, newsha)
	}

	return
}

func Parse(f string) (*Pluralfile, error) {
	pluralfile, err := os.Open(f)
	plrl := &Pluralfile{}
	if err != nil {
		return plrl, err
	}
	defer func(pluralfile *os.File) {
		_ = pluralfile.Close()
	}(pluralfile)

	scanner := bufio.NewScanner(pluralfile)
	r, _ := regexp.Compile(`^\s*$`)
	for scanner.Scan() {
		line := scanner.Text()
		ignore := r.MatchString(line)

		if ignore {
			continue
		}

		splitline, err := shlex.Split(line)
		if err != nil {
			return plrl, err
		}

		switch strings.ToLower(splitline[0]) {
		case "repo":
			plrl.Repo = splitline[1]
		case "helm":
			helms, err := expandGlob(splitline[1], func(targ string) Component {
				return &Helm{File: targ}
			})

			if err != nil {
				return plrl, err
			}

			plrl.Components = append(plrl.Components, helms...)
		case "tf":
			tfs, err := expandGlob(splitline[1], func(targ string) Component {
				return &Terraform{File: targ}
			})

			if err != nil {
				return plrl, err
			}

			plrl.Components = append(plrl.Components, tfs...)
		case "artifact":
			arts, err := expandGlob(splitline[1], func(targ string) Component {
				return &Artifact{File: targ, Platform: splitline[2], Arch: splitline[3]}
			})

			if err != nil {
				return plrl, err
			}

			plrl.Components = append(plrl.Components, arts...)
		case "ird":
			irds, err := expandGlob(splitline[1], func(targ string) Component {
				return &ResourceDefinition{File: targ}
			})

			if err != nil {
				return plrl, err
			}

			plrl.Components = append(plrl.Components, irds...)
		case "recipe":
			recipes, err := expandGlob(splitline[1], func(targ string) Component {
				return &Recipe{File: targ}
			})

			if err != nil {
				return plrl, err
			}

			plrl.Components = append(plrl.Components, recipes...)
		case "integration":
			integs, err := expandGlob(splitline[1], func(targ string) Component {
				return &Integration{File: targ}
			})

			if err != nil {
				return plrl, err
			}

			plrl.Components = append(plrl.Components, integs...)
		case "crd":
			chart := splitline[2]
			crds, err := expandGlob(splitline[1], func(targ string) Component {
				return &Crd{Chart: chart, File: targ}
			})

			if err != nil {
				return plrl, err
			}
			plrl.Components = append(plrl.Components, crds...)
		case "run":
			simpleSplit := strings.Split(line, " ")
			cmd, args := simpleSplit[1], simpleSplit[2:]
			plrl.Components = append(plrl.Components, &Command{Command: cmd, Args: args})
		case "tag":
			tags, err := expandGlob(splitline[1], func(tag string) Component {
				return &Tags{File: tag}
			})

			if err != nil {
				return plrl, err
			}
			plrl.Components = append(plrl.Components, tags...)
		case "attributes":
			pub, file := splitline[1], splitline[2]

			plrl.Components = append(plrl.Components, &RepoAttrs{File: file, Publisher: pub})
		default:
			continue
		}
	}

	return plrl, nil
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
