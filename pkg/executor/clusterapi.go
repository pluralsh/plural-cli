package executor

import (
	"os"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

func clusterAPISteps(path string) []*Step {
	pm, _ := manifest.FetchProject()
	sanitizedPath := pathing.SanitizeFilepath(path)

	homedir, _ := os.UserHomeDir()

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
		{
			Name:    "install bootstrap bounce",
			Wkdir:   sanitizedPath,
			Target:  pluralFile(path, "ONCE"),
			Command: "plural",
			Args:    []string{"--bootstrap", "wkspace", "helm", sanitizedPath, "--skip", "cluster-api-cluster", "--set", "cluster-api-operator.secret.bootstrap=true"},
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
			Args:    []string{"--bootstrap", "wkspace", "helm", sanitizedPath, "--set", "cluster-api-operator.secret.bootstrap=true"},
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
			Name:    "kube-init",
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
			Name:    "crds",
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
			Args:    []string{"wkspace", "helm", sanitizedPath, "--skip", "cluster-api-cluster", "--set", "cluster-api-operator.secret.bootstrap=true"},
			Sha:     "",
			Retries: 5,
		},
		{
			Name:    "clusterctl-move",
			Wkdir:   pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
			Target:  pluralFile(path, "ONCE"),
			Command: "clusterctl",
			Args:    []string{"move", "-n", "bootstrap", "--kubeconfig-context", "kind-bootstrap", "--to-kubeconfig", pathing.SanitizeFilepath(filepath.Join(homedir, ".kube", "config"))},
			Sha:     "",
		},
		{
			Name:    "delete bootstrap cluster",
			Target:  pluralFile(path, "ONCE"),
			Command: "plural",
			Args:    []string{"--bootstrap", "bootstrap", "cluster", "delete", "bootstrap"},
			Sha:     "",
		},
		// {
		// 	Name:    "terraform-init",
		// 	Wkdir:   pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
		// 	Target:  pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
		// 	Command: "terraform",
		// 	Args:    []string{"init", "-upgrade"},
		// 	Sha:     "",
		// },
		// {
		// 	Name:    "terraform-apply",
		// 	Wkdir:   pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
		// 	Target:  pathing.SanitizeFilepath(filepath.Join(path, "terraform")),
		// 	Command: "terraform",
		// 	Args:    []string{"apply", "-auto-approve"},
		// 	Sha:     "",
		// 	Retries: 2,
		// },
	}

	steps = append(steps, defaultSteps(path)...)

	return steps
}

func pluralFile(base, name string) string {
	return pathing.SanitizeFilepath(filepath.Join(base, ".plural", name))
}
