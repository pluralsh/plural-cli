package destroy

import (
	"os"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/executor"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

// TODO: where are normal destroys handled?
// TODO: the destroy for CAPI also needs to destroy the terraform. The question is when to do that?
// TODO: after destroy we need to nullify the deploy.hcl so on a new build it will go through bootstrapping again

func defaultDestroy(path string) []*executor.Step {
	pm, _ := manifest.FetchProject()
	sanitizedPath := pathing.SanitizeFilepath(path)
	homedir, _ := os.UserHomeDir()

	prov, _ := provider.GetProvider()

	clusterKubeContext := prov.KubeContext()

	// stateRemoveModuleArg := ""
	// switch pm.Provider {
	// case provider.AWS:
	// 	stateRemoveModuleArg = "module.aws-bootstrap-cluster-api.data.aws_eks_cluster.cluster"
	// case provider.GCP:
	// 	stateRemoveModuleArg = "module.gcp-bootstrap-cluster-api.data.google_container_cluster.cluster"
	// }

	providerBootstrapFlags := []string{}

	switch pm.Provider {
	case "aws":
		providerBootstrapFlags = []string{
			"--set", "cluster-api-provider-aws.cluster-api-provider-aws.bootstrapMode=true",
			"--set", "bootstrap.aws-ebs-csi-driver.enabled=false",
			"--set", "bootstrap.aws-load-balancer-controller.enabled=false",
			"--set", "bootstrap.cluster-autoscaler.enabled=false",
			"--set", "bootstrap.metrics-server.enabled=false",
			"--set", "bootstrap.snapshot-controller.enabled=false",
			"--set", "bootstrap.snapshot-validation-webhook.enabled=false",
			"--set", "bootstrap.tigera-operator.enabled=false",
		}
	case "azure":
		providerBootstrapFlags = []string{}
	case "gcp":
		providerBootstrapFlags = []string{}
	case "google":
		providerBootstrapFlags = []string{}
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
			Name:    "bootstrap crds",
			Wkdir:   sanitizedPath,
			Target:  pathing.SanitizeFilepath(filepath.Join(path, "crds")),
			Command: "plural",
			Args:    []string{"--bootstrap", "wkspace", "crds", sanitizedPath},
			Sha:     "",
		},
		{
			Name:    "install capi operators",
			Wkdir:   sanitizedPath,
			Target:  pluralFile(path, "ONCE"),
			Command: "plural",
			Args:    append([]string{"--bootstrap", "wkspace", "helm", sanitizedPath, "--skip", "cluster-api-cluster"}, providerBootstrapFlags...),
			Sha:     "",
			Retries: 2,
		},
		{
			Name:    "clusterctl-move",
			Wkdir:   pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Target:  pluralFile(path, "ONCE"),
			Command: "clusterctl", // TODO: we need to get the current context before we run these commands
			Args:    []string{"move", "-n", "bootstrap", "--kubeconfig-context", clusterKubeContext, "--to-kubeconfig", pathing.SanitizeFilepath(filepath.Join(homedir, ".kube", "config")), "--to-kubeconfig-context", "kind-bootstrap"},
			Sha:     "",
		},
		// {
		// 	Name:    "move",
		// 	Target:  pathing.SanitizeFilepath(path),
		// 	Command: "plural",
		// 	Args:    []string{"bootstrap", "cluster", "move"},
		// 	Sha:     "",
		// },
		// {
		// 	Name:    "bootstrap bounce",
		// 	Wkdir:   sanitizedPath,
		// 	Target:  pathing.SanitizeFilepath(filepath.Join(path, "helm")),
		// 	Command: "plural",
		// 	Args:    []string{"--bootstrap", "wkspace", "helm", sanitizedPath, "--skip", "cluster-api-cluster", "--set", "bootstrap-operator.operator.bootstrapMode=true"},
		// 	Sha:     "",
		// 	Retries: 2,
		// },
		// {
		// 	Name:    "progress",
		// 	Wkdir:   sanitizedPath,
		// 	Target:  pathing.SanitizeFilepath(filepath.Join(path, "helm")),
		// 	Command: "plural",
		// 	Args:    []string{"--bootstrap", "bootstrap", "cluster", "watch", "--wait-for-capi", pm.Cluster},
		// 	Sha:     "",
		// 	Retries: 1,
		// 	Verbose: true,
		// },
		{
			Name:    "wait-for-cluster",
			Wkdir:   sanitizedPath,
			Target:  pluralFile(path, "ONCE"),
			Command: "kubectl",
			Args:    []string{"wait", "--for=condition=ready", "-n", "bootstrap", "--timeout", "40m", "cluster", pm.Cluster}, // TODO: need to set the context to the bootstrap cluster
			Sha:     "",
		},
		{
			Name:    "wait-for-machines-running",
			Wkdir:   sanitizedPath,
			Target:  pluralFile(path, "ONCE"),
			Command: "kubectl",
			Args:    []string{"wait", "--for=jsonpath=.status.phase=Running", "-n", "bootstrap", "--timeout", "15m", "machinepool", "--all"}, // TODO: need to set the context to the bootstrap cluster
			Sha:     "",
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
		// {
		// 	Name:    "terraform-remove",
		// 	Wkdir:   pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
		// 	Target:  pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
		// 	Command: "terraform",
		// 	Args:    []string{"state", "rm", stateRemoveModuleArg},
		// 	Sha:     "",
		// },
	}
}

func pluralFile(base, name string) string {
	return pathing.SanitizeFilepath(filepath.Join(base, ".plural", name))
}
