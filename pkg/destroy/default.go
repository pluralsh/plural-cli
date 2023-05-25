package destroy

import (
	"path/filepath"

	"github.com/pluralsh/plural/pkg/provider"

	"github.com/pluralsh/plural/pkg/executor"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

func defaultDestroy(path string) []*executor.Step {
	pm, _ := manifest.FetchProject()
	sanitizedPath := pathing.SanitizeFilepath(path)
	stateRemoveModuleArg := ""
	switch pm.Provider {
	case provider.AWS:
		stateRemoveModuleArg = "module.aws-bootstrap-cluster-api.data.aws_eks_cluster.cluster"
	case provider.GCP:
		stateRemoveModuleArg = "module.gcp-bootstrap-cluster-api.data.google_container_cluster.cluster"
	}

	return []*executor.Step{
		{
			Name:    "create bootstrap cluster",
			Target:  pathing.SanitizeFilepath(path),
			Command: "plural",
			Args:    []string{"bootstrap", "cluster", "create", "bootstrap", "--skip-if-exists"},
			Sha:     "",
		},
		{
			Name:    "move",
			Target:  pathing.SanitizeFilepath(path),
			Command: "plural",
			Args:    []string{"bootstrap", "cluster", "move"},
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
			Args:    []string{"--bootstrap", "wkspace", "helm", sanitizedPath, "--skip", "cluster-api-cluster"},
			Sha:     "",
			Retries: 2,
		},
		{
			Name:    "progress",
			Wkdir:   sanitizedPath,
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "helm")),
			Command: "plural",
			Args:    []string{"--bootstrap", "bootstrap", "cluster", "watch", pm.Cluster},
			Sha:     "",
			Retries: 1,
			Verbose: true,
		},
		{
			Name:    "destroy cluster API",
			Wkdir:   sanitizedPath,
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "helm")),
			Command: "plural",
			Args:    []string{"bootstrap", "cluster", "destroy-cluster-api", pm.Cluster},
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
			Name:    "terraform-remove",
			Wkdir:   pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Command: "terraform",
			Args:    []string{"state", "rm", stateRemoveModuleArg},
			Sha:     "",
		},
	}
}
