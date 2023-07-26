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

func clusterAPIDeploySteps() []*Step {
	pm, _ := manifest.FetchProject()
	root, _ := git.Root()
	sanitizedPath := pathing.SanitizeFilepath(filepath.Join(root, "bootstrap"))

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
	case "google":
		providerBootstrapFlags = []string{
			"--set", "bootstrap.cert-manager.serviceAccount.create=true",
		}
	}

	return []*Step{
		{
			Name:       "build values",
			Args:       []string{"plural", "build", "--only", "bootstrap", "--force"},
			TargetPath: root,
			Execute:    RunPlural,
		},
		{
			Name:       "create bootstrap cluster",
			Args:       []string{"plural", "bootstrap", "cluster", "create", "bootstrap", "--skip-if-exists"},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "bootstrap crds",
			Args:       []string{"plural", "--bootstrap", "wkspace", "crds", "bootstrap"},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "install capi operators",
			Args:       append([]string{"plural", "--bootstrap", "wkspace", "helm", "bootstrap", "--skip", "cluster-api-cluster"}, providerBootstrapFlags...),
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "deploy cluster",
			Args:       append([]string{"plural", "--bootstrap", "wkspace", "helm", "bootstrap"}, providerBootstrapFlags...),
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "wait-for-cluster",
			Args:       []string{"plural", "--bootstrap", "clusters", "wait", "bootstrap", pm.Cluster},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "wait-for-machines-running",
			Args:       []string{"plural", "--bootstrap", "clusters", "mpwait", "bootstrap", pm.Cluster},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "init kubeconfig for target cluster",
			Args:       []string{"plural", "wkspace", "kube-init"},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "create-bootstrap-namespace-workload-cluster",
			Args:       []string{"plural", "bootstrap", "namespace", "create", "bootstrap"},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "install CRDs on target cluster",
			Args:       []string{"plural", "wkspace", "crds", "bootstrap"},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "clusterctl-init-workload",
			Args:       append([]string{"plural", "wkspace", "helm", "bootstrap", "--skip", "cluster-api-cluster"}, providerBootstrapFlags...),
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "clusterctl-move",
			Args:       []string{"plural", "bootstrap", "cluster", "move", "--kubeconfig-context", "kind-bootstrap", "--to-kubeconfig", pathing.SanitizeFilepath(filepath.Join(homedir, ".kube", "config"))},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "destroy kind cluster",
			Args:       []string{"plural", "--bootstrap", "bootstrap", "cluster", "delete", "bootstrap"},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
	}
}

func BootstrapClusterAPI() error {
	utils.Note("Bootstrapping cluster with Cluster API...\n")

	for _, step := range clusterAPIDeploySteps() {
		utils.Highlight("%s \n", step.Name)
		err := os.Chdir(step.TargetPath)
		if err != nil {
			return err
		}
		err = step.Execute(step.Args)
		if err != nil {
			return err
		}
	}

	utils.Success("Cluster bootstrapped successfully!")

	return nil
}
