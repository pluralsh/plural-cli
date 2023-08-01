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
	providerBootstrapFlags := getProviderBootstrapFlags(pm.Provider)
	prov, _ := provider.GetProvider()
	clusterKubeContext := prov.KubeContext()

	return []*Step{
		{
			Name:       "Create local bootstrap cluster",
			Args:       []string{"plural", "bootstrap", "cluster", "create", "bootstrap", "--skip-if-exists"},
			TargetPath: sanitizedPath,
			Execute:    RunPlural,
		},
		{
			Name:       "Bootstrap CRDs in local cluster",
			Args:       []string{"plural", "--bootstrap", "wkspace", "crds", "bootstrap"},
			TargetPath: sanitizedPath,
			Execute:    RunPlural,
		},
		{
			Name:       "Install Cluster API operators in local cluster",
			Args:       append([]string{"plural", "--bootstrap", "wkspace", "helm", "bootstrap", "--skip", "cluster-api-cluster"}, providerBootstrapFlags...),
			TargetPath: sanitizedPath,
			Execute:    RunPlural,
		},
		{
			Name:       "Move resources from target to local cluster",
			Args:       []string{"plural", "bootstrap", "cluster", "move", "--kubeconfig-context", clusterKubeContext, "--to-kubeconfig", pathing.SanitizeFilepath(filepath.Join(homedir, ".kube", "config")), "--to-kubeconfig-context", "kind-bootstrap"},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "Destroy bootstrap on target cluster",
			TargetPath: sanitizedPath,
			Execute: func(_ []string) error {
				return destroy()
			},
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
			Name:       "Cleanup cluster resources",
			Args:       nil,
			TargetPath: sanitizedPath,
			Execute:    CleanupClusterResources,
		},
		{
			Name:       "Destroy cluster API",
			Args:       []string{"plural", "bootstrap", "cluster", "destroy-cluster-api", pm.Cluster},
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

func CleanupClusterResources(_ []string) error {
	m, err := getMigrator()
	if err != nil {
		return err
	}

	return m.Destroy()
}
