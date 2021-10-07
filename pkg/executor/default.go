package executor

import (
	"path/filepath"
)

func defaultSteps(path string) []*Step {
	app := filepath.Base(path)

	return []*Step{
		{
			Name:    "terraform-init",
			Wkdir:   filepath.Join(path, "terraform"),
			Target:  filepath.Join(path, "terraform"),
			Command: "terraform",
			Args:    []string{"init"},
			Sha:     "",
		},
		{
			Name:    "terraform-apply",
			Wkdir:   filepath.Join(path, "terraform"),
			Target:  filepath.Join(path, "terraform"),
			Command: "terraform",
			Args:    []string{"apply", "-auto-approve"},
			Sha:     "",
			Retries: 1,
		},
		{
			Name:    "terraform-output",
			Wkdir:   app,
			Target:  filepath.Join(path, "terraform"),
			Command: "plural",
			Args:    []string{"output", "terraform", app},
			Sha:     "",
		},
		{
			Name:    "kube-init",
			Wkdir:   path,
			Target:  pluralfile(path, "NONCE"),
			Command: "plural",
			Args:    []string{"wkspace", "kube-init"},
			Sha:     "",
		},
		{
			Name:    "crds",
			Wkdir:   path,
			Target:  filepath.Join(path, "crds"),
			Command: "plural",
			Args:    []string{"wkspace", "crds", path},
			Sha:     "",
		},
		{
			Name:    "bounce",
			Wkdir:   path,
			Target:  filepath.Join(path, "helm"),
			Command: "plural",
			Args:    []string{"wkspace", "helm", path},
			Sha:     "",
			Retries: 1,
		},
	}
}
