package bootstrap

import (
	"os/exec"
	"path/filepath"

	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/backup"
)

func shouldDeleteProviderCluster(cluster, namespace string) bool {
	clusterExists := bootstrapClusterExists()
	deleting, err := IsClusterPhase(localClusterContext, cluster, namespace, capi.ClusterPhaseDeleting)

	if err != nil {
		return false
	}

	return clusterExists && !deleting
}

func shouldDeleteBootstrapCluster(cluster, namespace string) bool {
	clusterExists := bootstrapClusterExists()
	deleting, err := IsClusterPhase(localClusterContext, cluster, namespace, capi.ClusterPhaseDeleting)

	if err != nil {
		return false
	}

	return clusterExists && deleting
}

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

	clusterBackup := backup.NewCAPIBackup(man.Cluster)

	return []*Step{
		{
			Name:    "Destroy cluster API",
			Args:    []string{"plural", "bootstrap", "cluster", "destroy-cluster-api", man.Cluster},
			Execute: runPlural,
			Confirm: "It looks like your existing bootstrap cluster has a provider cluster configuration. All resources at your provider should be removed before continuing. Would you like to try and remove it automatically?",
			Skip:    !shouldDeleteProviderCluster(man.Cluster, "bootstrap"),
		},
		{
			Name:    "Destroy local bootstrap cluster",
			Args:    []string{"plural", "--bootstrap", "bootstrap", "cluster", "delete", "bootstrap"},
			Execute: runPlural,
			Confirm: "It looks like your existing bootstrap cluster has a provider cluster configuration in a non-recoverable state. Please make sure to manually delete all existing cluster resources at your provider before continuing. Would you like to destroy the bootstrap cluster?",
			Skip:    !shouldDeleteBootstrapCluster(man.Cluster, "bootstrap"),
		},
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
			Skip:    clusterBackup.Exists(),
		},
		{
			Name: "Restore cluster",
			Execute: func(_ []string) error {
				options := client.MoveOptions{
					ToKubeconfig: client.Kubeconfig{
						Path:    kubeconfigPath,
						Context: "kind-bootstrap",
					},
				}

				return clusterBackup.Restore(options)
			},
			Skip: !clusterBackup.Exists(),
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
			OnAfter: func() {
				options := client.MoveOptions{
					FromKubeconfig: client.Kubeconfig{
						Path:    kubeconfigPath,
						Context: "kind-bootstrap",
					},
				}

				err := clusterBackup.Save(options)
				if err != nil {
					_ = clusterBackup.Remove()
					utils.Error("error during saving state backup: %s", err)
				}
			},
		},
		{
			// TODO: Once https://github.com/kubernetes-sigs/cluster-api-provider-azure/issues/2498
			//  will be done we can use it and remove this step.
			Name: "Enable OIDC issuer",
			Execute: func(_ []string) error {
				return utils.Exec("az", "aks", "update", "-g", man.Project, "-n", man.Cluster, "--enable-oidc-issuer")
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
			Name:    "Destroy local bootstrap cluster",
			Args:    []string{"plural", "--bootstrap", "bootstrap", "cluster", "delete", "bootstrap"},
			Execute: runPlural,
			OnAfter: func() {
				err := clusterBackup.Remove()
				if err != nil {
					utils.Error("error during removing state backup: %s", err)
				}
			},
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
