package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/pluralsh/plural/pkg/bootstrap/azure"
	"github.com/pluralsh/plural/pkg/kubernetes"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"golang.org/x/oauth2/google"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// deleteSecrets deletes secrets matching label selector from given namespace in given context.
func deleteSecrets(context, namespace, labelSelector string) error {
	kubernetesClient, err := kubernetes.KubernetesWithContext(context)
	if err != nil {
		return err
	}

	return kubernetesClient.SecretDeleteCollection(namespace, meta.DeleteOptions{}, meta.ListOptions{LabelSelector: labelSelector})
}

// getSecrets returns secrets matching label selector from given namespace in given context.
func getSecrets(context, namespace, labelSelector string) (*v1.SecretList, error) {
	kubernetesClient, err := kubernetes.KubernetesWithContext(context)
	if err != nil {
		return nil, err
	}

	return kubernetesClient.SecretList(namespace, meta.ListOptions{LabelSelector: labelSelector})
}

// createSecrets creates secrets in given context.
func createSecrets(context string, secrets []v1.Secret) error {
	kubernetesClient, err := kubernetes.KubernetesWithContext(context)
	if err != nil {
		return err
	}

	for _, secret := range secrets {
		_, err := kubernetesClient.SecretCreate(secret.Namespace, prepareSecret(secret))
		if err != nil {
			return err
		}
	}

	return nil
}

// prepareSecret unsets read-only secret fields to prepare it for creation.
func prepareSecret(secret v1.Secret) *v1.Secret {
	secret.UID = ""
	secret.ResourceVersion = ""
	secret.Generation = 0
	secret.CreationTimestamp = meta.Time{}
	return &secret
}

// moveHelmSecrets moves secrets owned by Helm from one cluster to another.
// It requires source and target contexts in its arguments.
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
func getBootstrapFlags(prov string) []string {
	switch prov {
	case provider.AWS:
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
	case provider.AZURE:
		return []string{
			"--set", "cluster-api-cluster.cluster.azure.clusterIdentity.bootstrapMode=true",
		}
	case provider.GCP:
		return []string{
			"--set", "bootstrap.cert-manager.serviceAccount.create=true",
			"--set", "cluster-api-provider-gcp.cluster-api-provider-gcp.bootstrapMode=true",
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

// RunWithTempCredentials is a function wrapper that provides provider-specific flags with credentials
// that are used during bootstrap and destroy.
func RunWithTempCredentials(function ActionFunc) error {
	man, err := manifest.FetchProject()
	if err != nil {
		return err
	}

	var flags []string
	switch man.Provider {
	case provider.AZURE:
		as, err := azure.GetAuthService(utils.ToString(man.Context["SubscriptionId"]))
		if err != nil {
			return err
		}

		clientId, clientSecret, err := as.Setup(man.Cluster)
		if err != nil {
			return err
		}

		pathPrefix := "cluster-api-cluster.cluster.azure.clusterIdentity.bootstrapCredentials"
		flags = []string{
			"--set", fmt.Sprintf("%s.%s=%s", pathPrefix, "clientID", clientId),
			"--set", fmt.Sprintf("%s.%s=%s", pathPrefix, "clientSecret", clientSecret),
		}

		defer func(as *azure.AuthService) {
			err := as.Cleanup()
			if err != nil {
				utils.Error("%s", err)
			}
		}(as)
	case provider.AWS:
		ctx := context.Background()
		cfg, err := awsConfig.LoadDefaultConfig(ctx)
		if err != nil {
			return err
		}
		cred, err := cfg.Credentials.Retrieve(ctx)
		if err != nil {
			return err
		}
		pathPrefix := "cluster-api-provider-aws.cluster-api-provider-aws.managerBootstrapCredentials"
		flags = []string{
			"--set", fmt.Sprintf("%s.%s=%s", pathPrefix, "AWS_ACCESS_KEY_ID", cred.AccessKeyID),
			"--set", fmt.Sprintf("%s.%s=%s", pathPrefix, "AWS_SECRET_ACCESS_KEY", cred.SecretAccessKey),
			"--set", fmt.Sprintf("%s.%s=%s", pathPrefix, "AWS_SESSION_TOKEN", cred.SessionToken),
			"--set", fmt.Sprintf("%s.%s=%s", pathPrefix, "AWS_REGION", man.Region),
		}
	case provider.GCP:
		credentials, err := google.FindDefaultCredentials(context.Background())
		if err != nil {
			return err
		}
		flags = []string{
			"--setJSON", fmt.Sprintf(`cluster-api-provider-gcp.cluster-api-provider-gcp.managerBootstrapCredentials.credentialsJson=%q`, string(credentials.JSON)),
		}
	}

	return function(flags)
}
