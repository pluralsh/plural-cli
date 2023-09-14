package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"golang.org/x/oauth2/google"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/bootstrap/azure"
	"github.com/pluralsh/plural/pkg/kubernetes"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

var disableAzurePodIdentityFlag = []string{"--set", "bootstrap.azurePodIdentity.enabled=false"}

func applyManifest(manifest string) error {
	kube, err := kubernetes.Kubernetes()
	if err != nil {
		return err
	}

	f, err := os.CreateTemp("", "manifest")
	if err != nil {
		return err
	}
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			utils.Error("%s", err)
		}
	}(f.Name())

	_, err = f.WriteString(manifest)
	if err != nil {
		return err
	}

	return kube.Apply(f.Name(), true)
}

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
func moveHelmSecrets(sourceContext, targetContext string) error {
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
	case api.ProviderAWS:
		return []string{
			"--set", "cluster-api-provider-aws.cluster-api-provider-aws.bootstrapMode=true",
			"--set", "bootstrap.aws-ebs-csi-driver.enabled=false",
			"--set", "bootstrap.aws-load-balancer-controller.enabled=false",
			"--set", "bootstrap.cluster-autoscaler.enabled=false",
			"--set", "bootstrap.metrics-server.enabled=false",
			"--set", "bootstrap.snapshot-controller.enabled=false",
			"--set", "bootstrap.snapshot-validation-webhook.enabled=false",
			"--set", "bootstrap.tigera-operator.enabled=false",
			"--set", "bootstrap.external-dns.enabled=false",
			"--set", "plural-certmanager-webhook.enabled=false",
		}
	case api.ProviderAzure:
		return []string{
			"--set", "cluster-api-cluster.cluster.azure.clusterIdentity.bootstrapMode=true",
			"--set", "bootstrap.external-dns.enabled=false",
			"--set", "plural-certmanager-webhook.enabled=false",
		}
	case api.ProviderGCP:
		return []string{
			"--set", "bootstrap.cert-manager.serviceAccount.create=true",
			"--set", "cluster-api-provider-gcp.cluster-api-provider-gcp.bootstrapMode=true",
			"--set", "bootstrap.external-dns.enabled=false",
			"--set", "plural-certmanager-webhook.enabled=false",
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

func FilterSteps(steps []*Step) []*Step {
	filteredSteps := make([]*Step, 0, len(steps))
	for _, step := range steps {
		if !step.Skip {
			filteredSteps = append(filteredSteps, step)
		}
	}

	return filteredSteps
}

// ExecuteSteps of a bootstrap, migration or destroy process.
func ExecuteSteps(steps []*Step) error {
	defaultPath, err := GetBootstrapPath()
	if err != nil {
		return err
	}

	filteredSteps := FilterSteps(steps)
	for i, step := range filteredSteps {
		utils.Highlight("[%d/%d] %s\n", i+1, len(filteredSteps), step.Name)

		if step.SkipFunc != nil && step.SkipFunc() {
			utils.Highlight("Skipping step [%d/%d]\n", i+1, len(filteredSteps))
			continue
		}

		path := GetStepPath(step, defaultPath)
		err := os.Chdir(path)
		if err != nil {
			return err
		}

		for j := 0; j <= step.Retries; j++ {
			if j > 0 {
				utils.Highlight("Retrying, attempt %d of %d...\n", j, step.Retries)
			}
			err = step.Execute(step.Args)
			if err == nil {
				break
			}
			utils.Error("[%d/%d] %s failed: %s\n", i+1, len(filteredSteps), step.Name, err)
		}
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
	case api.ProviderAzure:
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
	case api.ProviderAWS:
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
	case api.ProviderGCP:
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
