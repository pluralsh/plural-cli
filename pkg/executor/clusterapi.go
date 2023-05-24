package executor

import (
	"path/filepath"

	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

func clusterAPISteps(path string) []*Step {
	pm, _ := manifest.FetchProject()
	sanitizedPath := pathing.SanitizeFilepath(path)

	return []*Step{
		{
			Name:    "create bootstrap cluster",
			Target:  pluralFile(path, "ONCE"),
			Command: "plural",
			Args:    []string{"bootstrap", "cluster", "create", "bootstrap", "--skip-if-exists"},
			Sha:     "",
		},
		{
			Name:    "bootstrap crds",
			Wkdir:   sanitizedPath,
			Target:  pluralFile(path, "ONCE"),
			Command: "plural",
			Args:    []string{"--bootstrap", "wkspace", "crds", sanitizedPath},
			Sha:     "",
		},
		{
			Name:    "bootstrap bounce",
			Wkdir:   sanitizedPath,
			Target:  pluralFile(path, "ONCE"),
			Command: "plural",
			Args:    []string{"--bootstrap", "wkspace", "helm", sanitizedPath, "--skip", "cluster-api-cluster"},
			Sha:     "",
			Retries: 2,
		},
		{
			Name:    "progress cluster API stack",
			Wkdir:   sanitizedPath,
			Target:  pluralFile(path, "ONCE"),
			Command: "plural",
			Args:    []string{"--bootstrap", "bootstrap", "cluster", "watch", pm.Cluster},
			Sha:     "",
			Retries: 1,
			Verbose: true,
		},
		{
			Name:    "cluster bounce",
			Wkdir:   sanitizedPath,
			Target:  pluralFile(path, "ONCE"),
			Command: "plural",
			Args:    []string{"--bootstrap", "wkspace", "helm", sanitizedPath},
			Sha:     "",
			Retries: 2,
		},
		{
			Name:    "progress cluster",
			Wkdir:   sanitizedPath,
			Target:  pluralFile(path, "ONCE"),
			Command: "plural",
			Args:    []string{"--bootstrap", "bootstrap", "cluster", "watch", "--move-cluster", pm.Cluster},
			Sha:     "",
			Retries: 1,
			Verbose: true,
		},
		{
			Name:    "delete bootstrap cluster",
			Target:  pluralFile(path, "ONCE"),
			Command: "plural",
			Args:    []string{"--bootstrap", "bootstrap", "cluster", "delete", "bootstrap"},
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
			Args:    []string{"wkspace", "helm", sanitizedPath, "--skip", "cluster-api-cluster"},
			Sha:     "",
			Retries: 2,
		},
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
	}
}

func pluralFile(base, name string) string {
	return pathing.SanitizeFilepath(filepath.Join(base, ".plural", name))
}
