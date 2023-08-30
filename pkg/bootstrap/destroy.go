package bootstrap

import (
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
)

// getDestroySteps returns list of steps to run during cluster destroy.
func getDestroySteps(destroy func() error, runPlural ActionFunc, additionalFlags []string) ([]*Step, error) {
	projectManifest, err := manifest.FetchProject()
	if err != nil {
		return nil, err
	}

	kubeconfigPath, err := getKubeconfigPath()
	if err != nil {
		return nil, err
	}

	flags := append(getBootstrapFlags(projectManifest.Provider), additionalFlags...)

	prov, err := provider.GetProvider()
	if err != nil {
		return nil, err
	}

	clusterKubeContext := prov.KubeContext()
	gitRootDir, err := git.Root()
	if err != nil {
		return nil, err
	}

	return []*Step{
		{
			Name:    "Create local bootstrap cluster",
			Args:    []string{"plural", "bootstrap", "cluster", "create", "bootstrap", "--skip-if-exists"},
			Execute: runPlural,
		},
		{
			Name:       "Rebuild values file",
			Args:       []string{"plural", "build-values", "bootstrap"},
			TargetPath: gitRootDir,
			Execute:    runPlural,
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
			Skip: func() bool {
				if _, err := CheckClusterReadiness(projectManifest.Cluster, "bootstrap"); err != nil {
					return true
				}

				return false
			},
		},
		{
			Name:    "Remove Helm secrets",
			Args:    []string{"kind-bootstrap"},
			Execute: removeHelmSecrets,
		},
		{
			Name:    "Move Helm secrets",
			Args:    []string{clusterKubeContext, "kind-bootstrap"},
			Execute: moveHelmSecrets,
		},
		{
			Name:    "Reinstall Helm charts to update configuration",
			Args:    append([]string{"plural", "--bootstrap", "wkspace", "helm", "bootstrap"}, flags...),
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
	}, nil
}

// DestroyCluster destroys cluster managed by Cluster API.
func DestroyCluster(destroy func() error, runPlural ActionFunc) error {
	utils.Highlight("Destroying Cluster API cluster...\n")

	if err := RunWithTempCredentials(func(flags []string) error {
		steps, err := getDestroySteps(destroy, runPlural, flags)
		if err != nil {
			return err
		}

		err = ExecuteSteps(steps)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	utils.Success("Cluster destroyed successfully!\n")
	return nil
}
