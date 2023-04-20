package executor

import (
	"fmt"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/provider"

	"github.com/pluralsh/plural/pkg/utils/pathing"
)

func clusterAPISteps(path string) []*Step {
	pm, _ := manifest.FetchProject()
	sanitizedPath := pathing.SanitizeFilepath(path)
	importModule := ""
	moduleArgs := ""

	// TODO: refactor
	switch pm.Provider {
	case provider.AWS:
		importModule = "module.aws-bootstrap-cluster-api.aws_eks_cluster.cluster"
		moduleArgs = pm.Cluster
	case provider.GCP:
		importModule = "module.google_container_cluster.cluster"
		moduleArgs = fmt.Sprintf("%s/%s/%s", pm.Project, pm.Region, pm.Cluster)

	}

	return []*Step{
		{
			Name:    "create bootstrap cluster",
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
			Args:    []string{"--bootstrap", "bootstrap", "cluster", "watch", "--enable-cluster-creation", pm.Cluster},
			Sha:     "",
			Retries: 1,
			Verbose: true,
		},
		{
			Name:    "delete bootstrap cluster",
			Target:  pathing.SanitizeFilepath(path),
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
			Args:    []string{"wkspace", "helm", sanitizedPath},
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
			Name:    "terraform-state-rm",
			Wkdir:   pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Command: "terraform",
			Args:    []string{"state", "rm", importModule},
			Sha:     "",
		},
		{
			Name:    "terraform-import",
			Wkdir:   pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Command: "terraform",
			Args:    []string{"import", importModule, moduleArgs},
			Sha:     "",
		},
	}
}
