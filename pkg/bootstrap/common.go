package bootstrap

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

// runTerraform executes terraform command with provided arguments, i.e. "terraform init".
func runTerraform(arguments []string) error {
	cmd := exec.Command("terraform", arguments...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// getBootstrapFlags returns list of provider-specific flags used during cluster bootstrap and destroy.
func getBootstrapFlags(provider string) []string {
	switch provider {
	case "aws":
		return []string{
			"--set", "cluster-api-provider-aws.cluster-api-provider-aws.bootstrapMode=true",
			"--set", "bootstrap.aws-ebs-csi-driver.enabled=false",
			"--set", "bootstrap.aws-load-balancer-controller.enabled=false",
			"--set", "bootstrap.cluster-autoscaler.enabled=false",
			"--set", "bootstrap.metrics-server.enabled=false",
			"--set", "bootstrap.snapshot-controller.enabled=false",
			"--set", "bootstrap.snapshot-validation-webhook.enabled=false",
			"--set", "bootstrap.tigera-operator.enabled=false",
		}
	case "google":
		return []string{
			"--set", "bootstrap.cert-manager.serviceAccount.create=true",
		}
	default:
		return []string{}
	}
}

// getKubeconfigPath returns path to kubeconfig in user home directory.
func getKubeconfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return pathing.SanitizeFilepath(filepath.Join(homeDir, ".kube", "config")), nil
}

// getBootstrapPath returns bootstrap repository path.
func getBootstrapPath() (string, error) {
	gitRootPath, err := git.Root()
	if err != nil {
		return "", err
	}

	return pathing.SanitizeFilepath(filepath.Join(gitRootPath, "bootstrap")), nil
}
