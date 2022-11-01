package executor

import (
	"fmt"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/manifest"
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

func defaultClusterAPISteps(path, provider string, manifest *manifest.ProjectManifest) []*Step {
	// app := filepath.Base(path)

	//TODO: ensure docker is installed and running
	//TODO: load environment variables needed for the provider -> Can be done through the clusterctl config file
	//TODO: find a way to handle syncing files/templates into the repo
	//TODO: find a way to only create kind cluster on initial deployment
	//TODO: deal with existing kind cluster
	//TODO: decide what namespace to use -> use terraform? this could then also be used for creating the files from templates

	//TODO: export  kubeconfig for deployed cluster
	//TODO: wait for cluster to be ready
	//TODO: move cluster management to deployed cluster
	//TODO: delete kind cluster
	return []*Step{
		{
			Name:    "kind-init",
			Wkdir:   filepath.Join(path, ""),
			Target:  filepath.Join(path, ""),
			Command: "kind",
			Args:    []string{"create", "cluster", "--wait", "300s", "--name", fmt.Sprintf("%s-bootstrap", manifest.Cluster)},
			Sha:     "",
		},
		{
			Name:    "clusterctl-init",
			Wkdir:   filepath.Join(path, ""),
			Target:  filepath.Join(path, ""),
			Command: "clusterctl",
			Args:    []string{"init", "--wait-providers", "--infrastructure", provider},
			Sha:     "",
		},
		{
			Name:    "create-bootstrap-namespace",
			Wkdir:   filepath.Join(path, ""),
			Target:  filepath.Join(path, ""),
			Command: "kubectl",
			Args:    []string{"create", "namespace", "bootstrap"},
			Sha:     "",
		},
		{
			Name:    "deploy-cluster",
			Wkdir:   filepath.Join(path, "terraform"),
			Target:  filepath.Join(path, "terraform"),
			Command: "kubectl",
			//TODO: remove hardcoded template
			Args: []string{"apply", "-f", "eqm-cluster-api.yaml"},
			Sha:  "",
		},
		{
			Name:    "wait-kubeconfig-secret",
			Wkdir:   filepath.Join(path, "terraform"),
			Target:  filepath.Join(path, "terraform"),
			Command: "/bin/sh",
			Args:    []string{"-c", fmt.Sprintf("while ! kubectl get secret -n bootstrap %s-kubeconfig; do sleep 1; done", manifest.Cluster)},
			Sha:     "",
		},
		{
			Name:    "get-kubeconfig",
			Wkdir:   filepath.Join(path, "terraform"),
			Target:  filepath.Join(path, "terraform"),
			Command: "/bin/sh",
			Args:    []string{"-c", fmt.Sprintf("clusterctl get kubeconfig -n bootstrap %s > kube_config_cluster.yaml", manifest.Cluster)},
			Sha:     "",
		},
		{
			Name:    "wait-for-cluster",
			Wkdir:   filepath.Join(path, ""),
			Target:  filepath.Join(path, ""),
			Command: "kubectl",
			Args:    []string{"wait", "--for=condition=ready", "-n", "bootstrap", "--timeout", "40m", "cluster", manifest.Cluster},
			Sha:     "",
		},
		{
			Name:    "wait-for-machines-running",
			Wkdir:   filepath.Join(path, ""),
			Target:  filepath.Join(path, ""),
			Command: "kubectl",
			Args:    []string{"wait", "--for=jsonpath=.status.phase=Running", "-n", "bootstrap", "--timeout", "15m", "machine", "--all"},
			Sha:     "",
		},
		//TODO: create bootstrap namespace on workload cluster. Use terraform? Difficult if terraform used earlier for templating.
		{
			Name:    "kube-init",
			Wkdir:   path,
			Target:  pluralfile(path, "NONCE"),
			Command: "plural",
			Args:    []string{"wkspace", "kube-init"},
			Sha:     "",
		},
		{
			//TODO: remove this testing command
			Name:    "deploy-cilium",
			Wkdir:   path,
			Target:  filepath.Join(path, ""),
			Command: "helm",
			Args:    []string{"install", "cilium", "cilium/cilium", "--version", "1.11.3", "--namespace", "bootstrap"},
			Sha:     "",
		},
		//TODO: this step requires a CNI to be installed. Perform Plural bootstrap helm install before continuing (also relevant for cert-manager)?
		{
			Name:    "clusterctl-init-workfload",
			Wkdir:   filepath.Join(path, "terraform"),
			Target:  filepath.Join(path, "terraform"),
			Command: "clusterctl",
			Args:    []string{"init", "--kubeconfig", "kube_config_cluster.yaml", "--wait-providers", "--infrastructure", provider},
			Sha:     "",
		},
		{
			Name:    "clusterctl-move",
			Wkdir:   filepath.Join(path, "terraform"),
			Target:  filepath.Join(path, "terraform"),
			Command: "clusterctl",
			Args:    []string{"move", "-n", "bootstrap", "--kubeconfig-context", fmt.Sprintf("kind-%s-bootstrap", manifest.Cluster), "--to-kubeconfig", "kube_config_cluster.yaml"},
			Sha:     "",
		},
		{
			Name:    "delete-kind-bootstrap-cluster",
			Wkdir:   filepath.Join(path, ""),
			Target:  filepath.Join(path, ""),
			Command: "kind",
			Args:    []string{"delete", "clusters", fmt.Sprintf("%s-bootstrap", manifest.Cluster)},
			Sha:     "",
		},
	}
}
