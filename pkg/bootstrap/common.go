package bootstrap

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

// removeHelmSecrets removes secrets owned by Helm from cluster bootstrap namespace.
func removeHelmSecrets(arguments []string) error {
	if len(arguments) != 1 {
		return fmt.Errorf("expected one context name in arguments, got %v instead", len(arguments))
	}

	context := arguments[0]

	cmd := exec.Command("kubectl", "delete", "secret", "-n", "bootstrap", "-l", "owner=helm", "--context", context)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// moveHelmSecrets moves secrets owned by Helm from one cluster to another.
func moveHelmSecrets(arguments []string) error {
	if len(arguments) != 2 {
		return fmt.Errorf("expected two context names in arguments, got %v instead", len(arguments))
	}

	sourceContext := arguments[0]
	targetContext := arguments[1]

	getCmd := exec.Command("kubectl", "--context", sourceContext, "get", "secret", "-n", "bootstrap", "-l", "owner=helm", "-o", "yaml")
	createCmd := exec.Command("kubectl", "--context", targetContext, "create", "-f", "-")

	r, w := io.Pipe()
	getCmd.Stdout = w
	getCmd.Stderr = os.Stderr
	createCmd.Stdin = r
	createCmd.Stdout = os.Stdout
	createCmd.Stderr = os.Stderr

	err := getCmd.Start()
	if err != nil {
		return err
	}

	err = createCmd.Start()
	if err != nil {
		return err
	}

	err = getCmd.Wait()
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	err = createCmd.Wait()
	if err != nil {
		return err
	}

	return err
}

// getEnvVar gets value of environment variable, if it is not set then default value is returned instead.
func getEnvVar(name, defaultValue string) string {
	if v, ok := os.LookupEnv(name); ok {
		return v
	}

	return defaultValue
}

// getBootstrapFlags returns list of provider-specific flags used during cluster bootstrap and destroy.
func getBootstrapFlags(provider string) []string {
	switch provider {
	case aws:
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
			"--set", "cluster-api-provider-gcp.cluster-api-provider-gcp.bootstrapMode=true",
		}
	case "azure":
		return []string{
			"--set", "cluster-api-cluster.cluster.azure.clusterIdentity.bootstrapMode=true",
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

// GetBootstrapPath returns bootstrap repository path.
func GetBootstrapPath() (string, error) {
	gitRootPath, err := git.Root()
	if err != nil {
		return "", err
	}

	return pathing.SanitizeFilepath(filepath.Join(gitRootPath, "bootstrap")), nil
}

// GetStepPath returns path from which step will be executed.
func GetStepPath(step *Step, defaultPath string) string {
	if step != nil && step.TargetPath != "" {
		return step.TargetPath
	}

	return defaultPath
}

// ExecuteSteps of a bootstrap, migration or destroy process.
func ExecuteSteps(steps []*Step) error {
	defaultPath, err := GetBootstrapPath()
	if err != nil {
		return err
	}

	for i, step := range steps {
		utils.Highlight("[%d/%d] %s \n", i+1, len(steps), step.Name)

		if step.Skip != nil && step.Skip() {
			continue
		}

		path := GetStepPath(step, defaultPath)
		err := os.Chdir(path)
		if err != nil {
			return err
		}

		err = step.Execute(step.Args)
		if err != nil {
			return err
		}
	}

	return nil
}

func RunWithTempCredentials(function ActionFunc) error {
	man, err := manifest.FetchProject()
	if err != nil {
		return err
	}

	prov, err := provider.GetProvider()
	if err != nil {
		return err
	}

	var flags []string

	switch man.Provider {
	case provider.AZURE:
		acs, err := GetAzureCredentialsService(utils.ToString(man.Context["SubscriptionId"]))
		if err != nil {
			return err
		}

		clientId, clientSecret, err := acs.Setup(man.Cluster)
		if err != nil {
			return err
		}

		pathPrefix := "cluster-api-cluster.cluster.azure.clusterIdentity.bootstrapCredentials"
		flags = []string{
			"--set", fmt.Sprintf("%s.%s=%s", pathPrefix, "clientID", clientId),
			"--set", fmt.Sprintf("%s.%s=%s", pathPrefix, "clientSecret", clientSecret),
		}

		defer func(acs *AzureCredentialsService) {
			err := acs.Cleanup()
			if err != nil {
				utils.Error("%s", err)
			}
		}(acs)
	case provider.GCP:
		credentials := prov.Context()["Credentials"]
		flags = []string{
			"--setJSON", fmt.Sprintf(`cluster-api-provider-gcp.cluster-api-provider-gcp.managerBootstrapCredentials.credentialsJson=%q`, credentials),
		}
	}

	return function(flags)
}
