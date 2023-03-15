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
			Name:    "namespace",
			Wkdir:   pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Command: "plural",
			Args:    []string{"bootstrap", "namespace", "create", app, "--skip-if-exists"},
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
