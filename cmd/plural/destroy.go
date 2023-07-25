package plural

import (
	"os"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/provider"

	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
)

func ExecuteClusterAPIDestroy(destroy func() error) error {
	root, err := git.Root()
	if err != nil {
		return err
	}
	bootstrapRepo := filepath.Join(root, "bootstrap")
	bootstrapRepoPath := pathing.SanitizeFilepath(bootstrapRepo)

	for _, step := range clusterAPIDestroySteps(bootstrapRepoPath, destroy) {
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

	return nil
}

func clusterAPIDestroySteps(path string, destroy func() error) []*Step {
	pm, _ := manifest.FetchProject()
	homedir, _ := os.UserHomeDir()
	sanitizedPath := pathing.SanitizeFilepath(path)
	providerBootstrapFlags := []string{}
	prov, _ := provider.GetProvider()
	clusterKubeContext := prov.KubeContext()

	runDestroy := func(_ []string) error {
		return destroy()
	}

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
			Name:       "create bootstrap cluster",
			Args:       []string{"plural", "bootstrap", "cluster", "create", "bootstrap", "--skip-if-exists"},
			TargetPath: sanitizedPath,
			Execute:    RunPlural,
		},
		{
			Name:       "bootstrap crds",
			Args:       []string{"plural", "--bootstrap", "wkspace", "crds", "bootstrap"},
			TargetPath: sanitizedPath,
			Execute:    RunPlural,
		},
		{
			Name:       "install capi operators",
			Args:       append([]string{"plural", "--bootstrap", "wkspace", "helm", "bootstrap", "--skip", "cluster-api-cluster"}, providerBootstrapFlags...),
			TargetPath: sanitizedPath,
			Execute:    RunPlural,
		},
		{
			Name:       "move",
			Args:       []string{"plural", "bootstrap", "cluster", "move", "--kubeconfig-context", clusterKubeContext, "--to-kubeconfig", pathing.SanitizeFilepath(filepath.Join(homedir, ".kube", "config")), "--to-kubeconfig-context", "kind-bootstrap"},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "destroy bootstrap on target cluster",
			TargetPath: sanitizedPath,
			Execute:    runDestroy,
		},
		{
			Name:       "wait for cluster",
			Args:       []string{"plural", "--bootstrap", "clusters", "wait", "bootstrap", pm.Cluster},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "wait for machines running",
			Args:       []string{"plural", "--bootstrap", "clusters", "mpwait", "bootstrap", pm.Cluster},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "destroy cluster API",
			Args:       []string{"plural", "bootstrap", "cluster", "destroy-cluster-api", pm.Cluster},
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
