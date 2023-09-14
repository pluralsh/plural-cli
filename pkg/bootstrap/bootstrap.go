package bootstrap

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
)

// getBootstrapSteps returns list of steps to run during cluster bootstrap.
func getBootstrapSteps(runPlural ActionFunc, additionalFlags []string) ([]*Step, error) {
	man, err := manifest.FetchProject()
	if err != nil {
		return nil, err
	}

	kubeconfigPath, err := getKubeconfigPath()
	if err != nil {
		return nil, err
	}

	bootstrapPath, err := GetBootstrapPath()
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
			Name:    "Deploy cluster",
			Args:    append([]string{"plural", "--bootstrap", "wkspace", "helm", "bootstrap"}, flags...),
			Execute: runPlural,
		},
		{
			Name:    "Wait for cluster",
			Args:    []string{"plural", "--bootstrap", "clusters", "wait", "bootstrap", man.Cluster},
			Execute: runPlural,
		},
		{
			Name: "Install Network",
			Execute: func(_ []string) error {
				return installCilium(man.Cluster)
			},
			Skip: man.Provider != api.ProviderKind,
		},
		{
			Name: "Install StorageClass",
			Execute: func(_ []string) error {
				return applyManifest(storageClassManifest)
			},
			Skip: man.Provider != api.ProviderKind,
		},
		{
			Name: "Save kubeconfig",
			Execute: func(_ []string) error {
				cmd := exec.Command("kind", "export", "kubeconfig", "--name", man.Cluster,
					"--kubeconfig", filepath.Join(bootstrapPath, "terraform", "kube_config_cluster.yaml"))
				return utils.Execute(cmd)
			},
			Skip: man.Provider != api.ProviderKind,
		},
		{
			Name:    "Wait for machine pools",
			Args:    []string{"plural", "--bootstrap", "clusters", "mpwait", "bootstrap", man.Cluster},
			Execute: runPlural,
		},
		{
			// TODO: Once https://github.com/kubernetes-sigs/cluster-api-provider-azure/issues/2498
			//  will be done we can use it and remove this step.
			Name: "Enable OIDC issuer",
			Execute: func(_ []string) error {
				cmd := exec.Command("az", "aks", "update", "-g", man.Project, "-n", man.Cluster, "--enable-oidc-issuer")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				return utils.Execute(cmd)
			},
			Skip: man.Provider != api.ProviderAzure,
		},
		{
			Name:    "Initialize kubeconfig for target cluster",
			Args:    []string{"plural", "wkspace", "kube-init"},
			Execute: runPlural,
		},
		{
			Name:    "Create bootstrap namespace in target cluster",
			Args:    []string{"plural", "bootstrap", "namespace", "create", "bootstrap"},
			Execute: runPlural,
		},
		{
			Name:    "Bootstrap CRDs in target cluster",
			Args:    []string{"plural", "wkspace", "crds", "bootstrap"},
			Execute: runPlural,
		},
		{
			Name:    "Install Cluster API operators in target cluster",
			Args:    append([]string{"plural", "wkspace", "helm", "bootstrap", "--skip", "cluster-api-cluster"}, append(flags, disableAzurePodIdentityFlag...)...),
			Execute: runPlural,
		},
		{
			Name:    "Move resources from local to target cluster",
			Args:    []string{"plural", "bootstrap", "cluster", "move", "--kubeconfig-context", localClusterContext, "--to-kubeconfig", kubeconfigPath},
			Execute: runPlural,
			Retries: 2,
		},
		{
			Name: "Move Helm secrets",
			Execute: func(_ []string) error {
				return moveHelmSecrets(localClusterContext, prov.KubeContext())
			},
			Retries: 2,
		},
		{
			Name:    "Destroy local cluster",
			Args:    []string{"plural", "--bootstrap", "bootstrap", "cluster", "delete", "bootstrap"},
			Execute: runPlural,
		},
	}, nil
}

// BootstrapCluster bootstraps cluster with Cluster API.
func BootstrapCluster(runPlural ActionFunc) error {
	utils.Highlight("Bootstrapping cluster with Cluster API...\n")

	if err := RunWithTempCredentials(func(flags []string) error {
		steps, err := getBootstrapSteps(runPlural, flags)
		if err != nil {
			return err
		}

		err = ExecuteSteps(steps)
		if err != nil {
			utils.Error("Cluster bootstrapping failed\n")
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	utils.Success("Cluster bootstrapped successfully!\n")
	return nil
}
