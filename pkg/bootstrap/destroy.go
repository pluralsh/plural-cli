package bootstrap

import (
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/provider"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

// getDestroySteps returns list of steps to run during cluster destroy.
func getDestroySteps(destroy func() error, runPlural ActionFunc, additionalFlags []string) ([]*Step, error) {
	man, err := manifest.FetchProject()
	if err != nil {
		return nil, err
	}

	kubeconfigPath, err := getKubeconfigPath()
	if err != nil {
		return nil, err
	}

	flags := append(getBootstrapFlags(man.Provider), additionalFlags...)

	prov, err := provider.GetProvider()
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
			Name:    "Bootstrap CRDs in local cluster",
			Args:    []string{"plural", "--bootstrap", "wkspace", "crds", "bootstrap"},
			Execute: runPlural,
		},
		{
			Name:    "Install Cluster API operators in local cluster",
			Args:    append([]string{"plural", "--bootstrap", "wkspace", "helm", "bootstrap", "--skip", "cluster-api-cluster"}, append(flags, disableAzurePodIdentityFlag...)...),
			Execute: runPlural,
		},
		{
			Name:    "Move resources from target to local cluster",
			Args:    []string{"plural", "bootstrap", "cluster", "move", "--kubeconfig-context", prov.KubeContext(), "--to-kubeconfig", kubeconfigPath, "--to-kubeconfig-context", localClusterContext},
			Execute: runPlural,
			SkipFunc: func() bool {
				_, err := CheckClusterReadiness(man.Cluster, "bootstrap")
				return err != nil
			},
			Retries: 2,
		},
		{
			Name: "Move Helm secrets",
			Execute: func(_ []string) error {
				return moveHelmSecrets(prov.KubeContext(), localClusterContext)
			},
			Retries: 2,
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
			Args:    []string{"plural", "--bootstrap", "clusters", "wait", "bootstrap", man.Cluster},
			Execute: runPlural,
		},
		{
			Name:    "Wait for machine pools",
			Args:    []string{"plural", "--bootstrap", "clusters", "mpwait", "bootstrap", man.Cluster},
			Execute: runPlural,
		},
		{
			Name:    "Destroy cluster API",
			Args:    []string{"plural", "bootstrap", "cluster", "destroy-cluster-api", man.Cluster},
			Execute: runPlural,
		},
		{
			Name:    "Destroy local bootstrap cluster",
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
