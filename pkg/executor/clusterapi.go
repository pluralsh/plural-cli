package executor

import (
	"os"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

func clusterAPISteps(path string) []*Step {
	pm, _ := manifest.FetchProject()
	// app := pathing.SanitizeFilepath(filepath.Base(path))
	sanitizedPath := pathing.SanitizeFilepath(path)

	homedir, _ := os.UserHomeDir()

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

	steps := []*Step{
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
		// {
		// 	Name:    "install prerequisites",
		// 	Wkdir:   sanitizedPath,
		// 	Target:  pluralFile(path, "ONCE"),
		// 	Command: "plural",
		// 	Args:    []string{"--bootstrap", "wkspace", "helm", sanitizedPath, "--skip", "cluster-api-operator", "--skip", "cluster-api-cluster", "--set", "cluster-api-operator.secret.bootstrap=true"},
		// 	Sha:     "",
		// 	Retries: 2,
		// },
		{
			Name:    "install capi operators",
			Wkdir:   sanitizedPath,
			Target:  pluralFile(path, "ONCE"),
			Command: "plural",
			Args:    append([]string{"--bootstrap", "wkspace", "helm", sanitizedPath, "--skip", "cluster-api-cluster"}, providerBootstrapFlags...),
			Sha:     "",
			Retries: 2,
		},
		// {
		// 	Name:    "progress cluster API stack",
		// 	Wkdir:   sanitizedPath,
		// 	Target:  pluralFile(path, "ONCE"),
		// 	Command: "plural",
		// 	Args:    []string{"--bootstrap", "bootstrap", "cluster", "watch", "--wait-for-capi", pm.Cluster},
		// 	Sha:     "",
		// 	Retries: 1,
		// 	Verbose: true,
		// },
		{
			Name:    "deploy cluster",
			Wkdir:   sanitizedPath,
			Target:  pluralFile(path, "ONCE"),
			Command: "plural",
			Args:    append([]string{"--bootstrap", "wkspace", "helm", sanitizedPath}, providerBootstrapFlags...),
			Sha:     "",
			Retries: 5,
		},
		{
			Name:    "wait-for-cluster",
			Wkdir:   sanitizedPath,
			Target:  pluralFile(path, "ONCE"),
			Command: "kubectl",
			Args:    []string{"wait", "--for=condition=ready", "-n", "bootstrap", "--timeout", "40m", "cluster", pm.Cluster},
			Sha:     "",
		},
		{
			Name:    "wait-for-machines-running",
			Wkdir:   sanitizedPath,
			Target:  pluralFile(path, "ONCE"),
			Command: "kubectl",
			Args:    []string{"wait", "--for=jsonpath=.status.phase=Running", "-n", "bootstrap", "--timeout", "15m", "machinepool", "--all"},
			Sha:     "",
		},
		// {
		// 	Name:    "progress cluster",
		// 	Wkdir:   sanitizedPath,
		// 	Target:  pluralFile(path, "ONCE"),
		// 	Command: "plural",
		// 	Args:    []string{"--bootstrap", "bootstrap", "cluster", "watch", pm.Cluster},
		// 	Sha:     "",
		// 	Retries: 1,
		// 	Verbose: true,
		// },
		{
			Name:    "kube-init-bootstrap",
			Wkdir:   sanitizedPath,
			Target:  pluralFile(path, "ONCE"),
			Command: "plural",
			Args:    []string{"wkspace", "kube-init"},
			Sha:     "",
		},
		{
			Name:    "create-bootstrap-namespace-workload-cluster",
			Wkdir:   sanitizedPath,
			Target:  pluralFile(path, "ONCE"),
			Command: "kubectl",
			Args:    []string{"create", "namespace", "bootstrap"},
			Sha:     "",
		},
		{
			Name:    "crds-bootstrap",
			Wkdir:   sanitizedPath,
			Target:  pluralFile(path, "ONCE"),
			Command: "plural",
			Args:    []string{"wkspace", "crds", sanitizedPath},
			Sha:     "",
		},
		{
			Name:    "clusterctl-init-workfload",
			Wkdir:   sanitizedPath,
			Target:  pluralFile(path, "ONCE"),
			Command: "plural",
			Args:    append([]string{"wkspace", "helm", sanitizedPath, "--skip", "cluster-api-cluster"}, providerBootstrapFlags...),
			Sha:     "",
			Retries: 5,
		},
		{
			Name:    "clusterctl-move",
			Wkdir:   pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Target:  pluralFile(path, "ONCE"),
			Command: "plural",
			Args:    []string{"bootstrap", "cluster", "move", "--kubeconfig-context", "kind-bootstrap", "--to-kubeconfig", pathing.SanitizeFilepath(filepath.Join(homedir, ".kube", "config"))},
			Sha:     "",
		},
		// { // TODO: re-anable this once we've debugged the move command so it works properly to avoid dangling resources
		// 	Name:    "delete bootstrap cluster",
		// 	Target:  pluralFile(path, "ONCE"),
		// 	Command: "plural",
		// 	Args:    []string{"--bootstrap", "bootstrap", "cluster", "delete", "bootstrap"},
		// 	Sha:     "",
		// },
	}

	steps = append(steps, defaultSteps(path)...)

	return steps
}

func pluralFile(base, name string) string {
	return pathing.SanitizeFilepath(filepath.Join(base, ".plural", name))
}
