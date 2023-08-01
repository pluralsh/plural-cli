package plural

import (
	"os"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

type ActionFunc func(arguments []string) error

type Step struct {
	Name             string
	Args             []string
	TargetPath       string
	BootstrapCommand bool
	Execute          ActionFunc
}

func getProviderBootstrapFlags(provider string) []string {
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
	default:
		return []string{}
	}
}

func clusterAPIDeploySteps() []*Step {
	pm, _ := manifest.FetchProject()
	root, _ := git.Root()
	sanitizedPath := pathing.SanitizeFilepath(filepath.Join(root, "bootstrap"))
	homedir, _ := os.UserHomeDir()
	providerBootstrapFlags := getProviderBootstrapFlags(pm.Provider)

	return []*Step{
		{
			Name:       "Create local bootstrap cluster",
			Args:       []string{"plural", "bootstrap", "cluster", "create", "bootstrap", "--skip-if-exists"},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "Bootstrap CRDs in local cluster",
			Args:       []string{"plural", "--bootstrap", "wkspace", "crds", "bootstrap"},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "Install Cluster API operators in local cluster",
			Args:       append([]string{"plural", "--bootstrap", "wkspace", "helm", "bootstrap", "--skip", "cluster-api-cluster"}, providerBootstrapFlags...),
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "Deploy cluster",
			Args:       append([]string{"plural", "--bootstrap", "wkspace", "helm", "bootstrap"}, providerBootstrapFlags...),
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "Wait for cluster",
			Args:       []string{"plural", "--bootstrap", "clusters", "wait", "bootstrap", pm.Cluster},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "Wait for machine pools",
			Args:       []string{"plural", "--bootstrap", "clusters", "mpwait", "bootstrap", pm.Cluster},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "Initialize kubeconfig for target cluster",
			Args:       []string{"plural", "wkspace", "kube-init"},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "Create bootstrap namespace in target cluster",
			Args:       []string{"plural", "bootstrap", "namespace", "create", "bootstrap"},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "Bootstrap CRDs in target cluster",
			Args:       []string{"plural", "wkspace", "crds", "bootstrap"},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "Install Cluster API operators in target cluster",
			Args:       append([]string{"plural", "wkspace", "helm", "bootstrap", "--skip", "cluster-api-cluster"}, providerBootstrapFlags...),
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "Move resources from local to target cluster",
			Args:       []string{"plural", "bootstrap", "cluster", "move", "--kubeconfig-context", "kind-bootstrap", "--to-kubeconfig", pathing.SanitizeFilepath(filepath.Join(homedir, ".kube", "config"))},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "Destroy local cluster",
			Args:       []string{"plural", "--bootstrap", "bootstrap", "cluster", "delete", "bootstrap"},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
	}
}

func BootstrapClusterAPI() error {
	utils.Highlight("Bootstrapping cluster with Cluster API...\n")

	steps := clusterAPIDeploySteps()
	for i, step := range steps {
		utils.Highlight("[%d/%d] %s \n", i+1, len(steps), step.Name)
		err := os.Chdir(step.TargetPath)
		if err != nil {
			return err
		}
		err = step.Execute(step.Args)
		if err != nil {
			return err
		}
	}

	utils.Success("Cluster bootstrapped successfully!\n")
	return nil
}
