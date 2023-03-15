package executor

import (
	"path/filepath"

	"github.com/pluralsh/plural/pkg/utils/pathing"
)

func clusterAPISteps(path string) []*Step {
	//app := pathing.SanitizeFilepath(filepath.Base(path))
	sanitizedPath := pathing.SanitizeFilepath(path)

	return []*Step{
		{
			Name:    "create-bootstrap-cluster",
			Target:  pathing.SanitizeFilepath(path),
			Command: "plural",
			Args:    []string{"bootstrap", "cluster", "create", "bootstrap", "--skip-if-exists"},
			Sha:     "",
		},
		{
			Name:    "bootstrap crds",
			Wkdir:   sanitizedPath,
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "crds")),
			Command: "plural",
			Args:    []string{"--bootstrap", "wkspace", "crds", sanitizedPath},
			Sha:     "",
		},
		{
			Name:    "bootstrap bounce",
			Wkdir:   sanitizedPath,
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "helm")),
			Command: "plural",
			Args:    []string{"--bootstrap", "wkspace", "helm", sanitizedPath},
			Sha:     "",
			Retries: 2,
		},
		{
			Name:    "progress",
			Wkdir:   sanitizedPath,
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "helm")),
			Command: "plural",
			Args:    []string{"--bootstrap", "bootstrap", "cluster", "watch", "test-aws"},
			Sha:     "",
			Retries: 1,
			Verbose: true,
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
