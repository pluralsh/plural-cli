package bootstrap

import (
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

	flags := getBootstrapFlags(projectManifest.Provider)

	prov, err := provider.GetProvider()
	if err != nil {
		return nil, err
	}

	clusterKubeContext := prov.KubeContext()
	var steps []*Step

	steps = append(steps, []*Step{
		{
			Name:    "Create local bootstrap cluster",
			Args:    []string{"plural", "bootstrap", "cluster", "create", "bootstrap", "--skip-if-exists"},
			Execute: runPlural,
		},
		{
			Name:    "Bootstrap CRDs in local cluster",
			Args:    []string{"plural", "--bootstrap", "wkspace", "crds", "bootstrap"},
			Execute: runPlural,
		},
		{
			Name:    "Install Cluster API operators in local cluster",
			Args:    append([]string{"plural", "--bootstrap", "wkspace", "helm", "bootstrap", "--skip", "cluster-api-cluster"}, flags...),
			Execute: runPlural,
		},
		{
			Name:    "Move resources from target to local cluster",
			Args:    []string{"plural", "bootstrap", "cluster", "move", "--kubeconfig-context", clusterKubeContext, "--to-kubeconfig", kubeconfigPath, "--to-kubeconfig-context", "kind-bootstrap"},
			Execute: runPlural,
		},
		{
			Name: "Destroy bootstrap on target cluster",
			Execute: func(_ []string) error {
				return destroy()
			},
		},
		{
			Name:    "Wait for cluster",
			Args:    []string{"plural", "--bootstrap", "clusters", "wait", "bootstrap", projectManifest.Cluster},
			Execute: runPlural,
		},
	}...)
	if projectManifest.Provider == "kind" {
		steps = append(steps, []*Step{
			{
				Name: "Install Network",
				Execute: func(_ []string) error {
					return InstallCilium(projectManifest.Cluster)
				},
			},
		}...)
	}
	steps = append(steps, []*Step{
		{
			Name:    "Wait for machine pools",
			Args:    []string{"plural", "--bootstrap", "clusters", "mpwait", "bootstrap", projectManifest.Cluster},
			Execute: runPlural,
		},
		{
			Name: "Cleanup cluster resources",
			Execute: func(_ []string) error {
				m, err := getMigrator()
				if err != nil {
					return err
				}

				return m.Destroy()
			},
		},
		{
			Name:    "Destroy cluster API",
			Args:    []string{"plural", "bootstrap", "cluster", "destroy-cluster-api", projectManifest.Cluster},
			Execute: runPlural,
		},
		{
			Name:    "Destroy local cluster",
			Args:    []string{"plural", "--bootstrap", "bootstrap", "cluster", "delete", "bootstrap"},
			Execute: runPlural,
		},
	}...)
	return steps, nil
}

// DestroyCluster destroys cluster managed by Cluster API.
func DestroyCluster(destroy func() error, runPlural ActionFunc) error {
	utils.Highlight("Destroying Cluster API cluster...\n")

	steps, err := getDestroySteps(destroy, runPlural)
	if err != nil {
		return err
	}

	err = ExecuteSteps(steps)
	if err != nil {
		return err
	}

	utils.Success("Cluster destroyed successfully!\n")
	return nil
}
