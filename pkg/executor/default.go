package executor

import (
	"path/filepath"
)

func defaultSteps(path string) []*Step {
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
		},
		{
			Name:    "kube-init",
			Wkdir:   path,
			Target:  forgefile(path, "NONCE"),
			Command: "forge",
			Args:    []string{"wkspace", "kube-init", path},
			Sha:     "",
		},
		{
			Name:    "docker-credentials",
			Wkdir:   path,
			Target:  forgefile(path, "ONCE"),
			Command: "forge",
			Args:    []string{"wkspace", "docker-credentials", path},
			Sha:     "",
		},
		{
			Name:    "crds",
			Wkdir:   path,
			Target:  filepath.Join(path, "crds"),
			Command: "forge",
			Args:    []string{"wkspace", "crds", path},
			Sha:     "",
		},
		{
			Name:    "bounce",
			Wkdir:   path,
			Target:  filepath.Join(path, "helm"),
			Command: "forge",
			Args:    []string{"wkspace", "helm", path},
			Sha:     "",
		},
	}
}