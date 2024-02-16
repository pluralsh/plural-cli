package utils

import (
	"os/exec"
	"strings"
)

const noKindClustersError = "No kind clusters found."

func IsKindClusterAlreadyExists(name string) bool {
	cmd := exec.Command("kind", "get", "clusters")
	out, err := ExecuteWithOutput(cmd)
	if err != nil {
		return false
	}
	if strings.Contains(out, noKindClustersError) {
		return false
	}

	return strings.Contains(out, name)
}

func GetKindClusterKubeconfig(name string, internal bool) (string, error) {
	kubeconfigArgs := []string{"get", "kubeconfig", "--name", name}
	if internal {
		kubeconfigArgs = append(kubeconfigArgs, "--internal")
	}

	cmd := exec.Command("kind", kubeconfigArgs...)
	return ExecuteWithOutput(cmd)
}
