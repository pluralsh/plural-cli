package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/kubernetes"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteSecrets(context, namespace string, labelSelector string) error {
	k, err := kubernetes.KubernetesWithContext(context)
	if err != nil {
		return err
	}

	return k.SecretDeleteCollection(namespace, meta.DeleteOptions{}, meta.ListOptions{LabelSelector: labelSelector})
}

func getSecrets(context, namespace, labelSelector string) (*v1.SecretList, error) {
	k, err := kubernetes.KubernetesWithContext(context)
	if err != nil {
		return nil, err
	}

	return k.SecretList(namespace, meta.ListOptions{LabelSelector: labelSelector})
}

func createSecrets(context string, secrets []v1.Secret) error {
	k, err := kubernetes.KubernetesWithContext(context)
	if err != nil {
		return err
	}

	for _, secret := range secrets {
		_, err := k.SecretCreate(secret.Namespace, prepareSecret(secret))
		if err != nil {
			return err
		}
	}

	return nil
}

func prepareSecret(secret v1.Secret) *v1.Secret {
	secret.UID = ""
	secret.ResourceVersion = ""
	secret.Generation = 0
	secret.CreationTimestamp = meta.Time{}
	return &secret
}

// moveHelmSecrets moves secrets owned by Helm from one cluster to another.
func moveHelmSecrets(arguments []string) error {
	if len(arguments) != 2 {
		return fmt.Errorf("expected two context names in arguments, got %v instead", len(arguments))
	}
	sourceContext := arguments[0]
	targetContext := arguments[1]

	err := deleteSecrets(targetContext, "bootstrap", "owner=helm")
	if err != nil {
		return err
	}

	secrets, err := getSecrets(sourceContext, "bootstrap", "owner=helm")
	if err != nil {
		return err
	}

	return createSecrets(targetContext, secrets.Items)
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
	}

	return function(flags)
}
