package executor

import (
	"path/filepath"

	"github.com/pluralsh/plural/pkg/utils/pathing"
)

func defaultSteps(path string) []*Step {
	app := pathing.SanitizeFilepath(filepath.Base(path))
	sanitizedPath := pathing.SanitizeFilepath(path)

	return []*Step{
		{
			Name:    "terraform-init",
			Wkdir:   pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Command: "terraform",
			Args:    []string{"init", "-upgrade"},
			Sha:     "",
		},
		{
			Name:    "terraform-apply",
			Wkdir:   pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Command: "terraform",
			Args:    []string{"apply", "-auto-approve"},
			Sha:     "",
			Retries: 2,
		},
		{
			Name:    "terraform-output",
			Wkdir:   app,
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Command: "plural",
			Args:    []string{"output", "terraform", app},
			Sha:     "",
		},
		{
			Name:    "kube-init",
			Wkdir:   sanitizedPath,
			Target:  pathing.SanitizeFilepath(pluralfile(path, "NONCE")),
			Command: "plural",
			Args:    []string{"wkspace", "kube-init"},
			Sha:     "",
		},
		{
			Name:    "crds",
			Wkdir:   sanitizedPath,
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "crds")),
			Command: "plural",
			Args:    []string{"wkspace", "crds", sanitizedPath},
			Sha:     "",
		},
		{
			Name:    "bounce",
			Wkdir:   sanitizedPath,
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "helm")),
			Command: "plural",
			Args:    []string{"wkspace", "helm", sanitizedPath},
			Sha:     "",
			Retries: 2,
		},
	}
}
