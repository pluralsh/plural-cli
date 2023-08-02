package bootstrap

import (
	"os"

	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
)

// getDestroySteps returns list of steps to run during cluster destroy.
func getDestroySteps(destroy func() error, runPlural ActionFunc) ([]*Step, error) {
	projectManifest, err := manifest.FetchProject()
	if err != nil {
		return nil, err
	}

	kubeconfigPath, err := getKubeconfigPath()
	if err != nil {
		return nil, err
	}

	bootstrapPath, err := getBootstrapPath()
	if err != nil {
		return nil, err
	}

	flags := getBootstrapFlags(projectManifest.Provider)

	prov, err := provider.GetProvider()
	if err != nil {
		return nil, err
	}

	clusterKubeContext := prov.KubeContext()

	return []*Step{
		{
			Name:       "Create local bootstrap cluster",
			Args:       []string{"plural", "bootstrap", "cluster", "create", "bootstrap", "--skip-if-exists"},
			TargetPath: bootstrapPath,
			Execute:    runPlural,
		},
		{
			Name:       "Bootstrap CRDs in local cluster",
			Args:       []string{"plural", "--bootstrap", "wkspace", "crds", "bootstrap"},
			TargetPath: bootstrapPath,
			Execute:    runPlural,
		},
		{
			Name:       "Install Cluster API operators in local cluster",
			Args:       append([]string{"plural", "--bootstrap", "wkspace", "helm", "bootstrap", "--skip", "cluster-api-cluster"}, flags...),
			TargetPath: bootstrapPath,
			Execute:    runPlural,
		},
		{
			Name:       "Move resources from target to local cluster",
			Args:       []string{"plural", "bootstrap", "cluster", "move", "--kubeconfig-context", clusterKubeContext, "--to-kubeconfig", kubeconfigPath, "--to-kubeconfig-context", "kind-bootstrap"},
			Execute:    runPlural,
			TargetPath: bootstrapPath,
		},
		{
			Name:       "Destroy bootstrap on target cluster",
			TargetPath: bootstrapPath,
			Execute: func(_ []string) error {
				return destroy()
			},
		},
		{
			Name:       "Wait for cluster",
			Args:       []string{"plural", "--bootstrap", "clusters", "wait", "bootstrap", projectManifest.Cluster},
			Execute:    runPlural,
			TargetPath: bootstrapPath,
		},
		{
			Name:       "Wait for machine pools",
			Args:       []string{"plural", "--bootstrap", "clusters", "mpwait", "bootstrap", projectManifest.Cluster},
			Execute:    runPlural,
			TargetPath: bootstrapPath,
		},
		{
			Name:       "Cleanup cluster resources",
			TargetPath: bootstrapPath,
			Execute: func(_ []string) error {
				m, err := getMigrator()
				if err != nil {
					return err
				}

				return m.Destroy()
			},
		},
		{
			Name:       "Destroy cluster API",
			Args:       []string{"plural", "bootstrap", "cluster", "destroy-cluster-api", projectManifest.Cluster},
			Execute:    runPlural,
			TargetPath: bootstrapPath,
		},
		{
			Name:       "Destroy local cluster",
			Args:       []string{"plural", "--bootstrap", "bootstrap", "cluster", "delete", "bootstrap"},
			Execute:    runPlural,
			TargetPath: bootstrapPath,
		},
	}, nil
}

// DestroyCluster destroys cluster managed by Cluster API.
func DestroyCluster(destroy func() error, runPlural ActionFunc) error {

	utils.Highlight("Destroying Cluster API cluster...\n")

	steps, err := getDestroySteps(destroy, runPlural)
	if err != nil {
		return err
	}

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

	utils.Success("Cluster destroyed successfully!\n")
	return nil
}
