package bootstrap

import (
	"os/exec"
	"path/filepath"

	"sigs.k8s.io/cluster-api/cmd/clusterctl/client"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/capi"
)

// saveKindKubeconfig exports kind kubeconfig to file.
func saveKindKubeconfig(_ []string) error {
	man, err := manifest.FetchProject()
	if err != nil {
		return err
	}

	bootstrapPath, err := GetBootstrapPath()
	if err != nil {
		return err
	}

	cmd := exec.Command("kind", "export", "kubeconfig", "--name", man.Cluster,
		"--kubeconfig", filepath.Join(bootstrapPath, "terraform", "kube_config_cluster.yaml"))
	return utils.Execute(cmd)
}

func enableAzureOIDCIssuer(_ []string) error {
	man, err := manifest.FetchProject()
	if err != nil {
		return err
	}

	cmd := exec.Command("az", "aks", "update", "-g", man.Project, "-n", man.Cluster, "--enable-oidc-issuer")
	return utils.Execute(cmd)
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
			Skip:    capi.MoveBackupExists(),
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

				return capi.RestoreMoveBackup(options)
			},
			Skip: !capi.MoveBackupExists(),
		},
		{
			Name:    "Wait for cluster",
			Args:    []string{"plural", "--bootstrap", "clusters", "wait", "bootstrap", man.Cluster},
			Execute: runPlural,
		},
		{
			Name: "Install Network",
			Execute: func(_ []string) error {
				return InstallCilium(man.Cluster)
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
			Name:    "Save kubeconfig",
			Execute: saveKindKubeconfig,
			Skip:    man.Provider != api.ProviderKind,
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

				err := capi.SaveMoveBackup(options)
				if err != nil {
					capi.RemoveStateBackup()
					utils.Error("error during saving state backup: %s", err)
				}
			},
		},
		{
			// TODO: Once https://github.com/kubernetes-sigs/cluster-api-provider-azure/issues/2498
			//  will be done we can use it and remove this step.
			Name:    "Enable OIDC issuer",
			Execute: enableAzureOIDCIssuer,
			Skip:    man.Provider != api.ProviderAzure,
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
			Args:    []string{"plural", "bootstrap", "cluster", "move", "--kubeconfig-context", "kind-bootstrap", "--to-kubeconfig", kubeconfigPath},
			Execute: runPlural,
			Retries: 2,
		},
		{
			Name: "Move Helm secrets",
			Execute: func(_ []string) error {
				return moveHelmSecrets("kind-bootstrap", prov.KubeContext())
			},
			Retries: 2,
		},
		{
			Name:    "Destroy local cluster",
			Args:    []string{"plural", "--bootstrap", "bootstrap", "cluster", "delete", "bootstrap"},
			Execute: runPlural,
			OnAfter: func() {
				err := capi.RemoveStateBackup()
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
