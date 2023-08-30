package bootstrap

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/kubernetes"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/provider"
	"github.com/pluralsh/plural/pkg/utils"
)

// getBootstrapSteps returns list of steps to run during cluster bootstrap.
func getBootstrapSteps(runPlural ActionFunc, additionalFlags []string) ([]*Step, error) {
	projectManifest, err := manifest.FetchProject()
	if err != nil {
		return nil, err
	}

	kubeconfigPath, err := getKubeconfigPath()
	if err != nil {
		return nil, err
	}

	flags := append(getBootstrapFlags(projectManifest.Provider), additionalFlags...)

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
			Args:    append([]string{"plural", "--bootstrap", "wkspace", "helm", "bootstrap", "--skip", "cluster-api-cluster"}, flags...),
			Execute: runPlural,
		},
		{
			Name:    "Deploy cluster",
			Args:    append([]string{"plural", "--bootstrap", "wkspace", "helm", "bootstrap"}, flags...),
			Execute: runPlural,
		},
		{
			Name:    "Wait for cluster",
			Args:    []string{"plural", "--bootstrap", "clusters", "wait", "bootstrap", projectManifest.Cluster},
			Execute: runPlural,
		},

		/* ====== START - KIND STEPS ====== */
		{
			Name: "Install Network",
			Execute: func(_ []string) error {
				return InstallCilium(projectManifest.Cluster)
			},
			Skip: func() bool {
				return projectManifest.Provider != provider.KIND
			},
		},
		{
			Name: "Install StorageClass",
			Execute: func(_ []string) error {
				kube, err := kubernetes.Kubernetes()
				if err != nil {
					return err
				}
				f, err := os.CreateTemp("", "storageClass")
				if err != nil {
					return err
				}
				defer os.Remove(f.Name())
				_, err = f.WriteString(storageClassManifest)
				if err != nil {
					return err
				}
				if err := kube.Apply(f.Name(), true); err != nil {
					return err
				}

				return nil
			},
			Skip: func() bool {
				return projectManifest.Provider != provider.KIND
			},
		},
		{
			Name: "Save kubeconfig",
			Execute: func(_ []string) error {
				bootstrapPath, err := GetBootstrapPath()
				if err != nil {
					return err
				}
				cmd := exec.Command("kind", "export", "kubeconfig", "--name", projectManifest.Cluster, "--kubeconfig", filepath.Join(bootstrapPath, "terraform", "kube_config_cluster.yaml"))
				if err := utils.Execute(cmd); err != nil {
					return err
				}

				return nil
			},
			Skip: func() bool {
				return projectManifest.Provider != provider.KIND
			},
		},
		/* ====== END - KIND STEPS ====== */

		{
			Name:    "Wait for machine pools",
			Args:    []string{"plural", "--bootstrap", "clusters", "mpwait", "bootstrap", projectManifest.Cluster},
			Execute: runPlural,
		},

		/* ====== START - AZURE STEPS ====== */
		// TODO:
		//  Once https://github.com/kubernetes-sigs/cluster-api-provider-azure/issues/2498
		//  will be done we can use it and remove this step.
		{
			Name: "Enable OIDC issuer",
			Execute: func(_ []string) error {
				cmd := exec.Command("az", "aks", "update", "-g", projectManifest.Project, "-n", projectManifest.Cluster, "--enable-oidc-issuer")
				if err := utils.Execute(cmd); err != nil {
					return err
				}

				return nil
			},
			Skip: func() bool {
				return projectManifest.Provider != provider.AZURE
			},
		},
		/* ====== END - AZURE STEPS ====== */

		{
			Name: "Post install resources",
			Execute: func(_ []string) error {
				m, err := getMigrator()
				if err != nil {
					return err
				}

				return m.PostInstall()
			},
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
			Args:    append([]string{"plural", "wkspace", "helm", "bootstrap", "--skip", "cluster-api-cluster"}, flags...),
			Execute: runPlural,
		},
		{
			Name:    "Move resources from local to target cluster",
			Args:    []string{"plural", "bootstrap", "cluster", "move", "--kubeconfig-context", "kind-bootstrap", "--to-kubeconfig", kubeconfigPath},
			Execute: runPlural,
		},
		{
			Name: "Remove Helm secrets",
			Execute: func(arguments []string) error {
				cmd := exec.Command("kubectl", "delete", "secret", "-n", "bootstrap", "-l", "owner=helm")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				return cmd.Run()
			},
		},
		{
			Name: "Copy Helm secrets",
			Execute: func(arguments []string) error {
				getCmd := exec.Command("kubectl", "get", "secret", "-n", "bootstrap", "-l", "owner=helm", "-o", "yaml", "--context", "kind-bootstrap")
				createCmd := exec.Command("kubectl", "create", "-f", "-")

				r, w := io.Pipe()
				getCmd.Stdout = w
				getCmd.Stderr = os.Stderr
				createCmd.Stdin = r
				createCmd.Stdout = os.Stdout
				createCmd.Stderr = os.Stderr

				err := getCmd.Start()
				if err != nil {
					return err
				}

				err = createCmd.Start()
				if err != nil {
					return err
				}

				err = getCmd.Wait()
				if err != nil {
					return err
				}

				err = w.Close()
				if err != nil {
					return err
				}

				err = createCmd.Wait()
				if err != nil {
					return err
				}

				return err
			},
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
			deleteBootstrapCluster(runPlural)
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	utils.Success("Cluster bootstrapped successfully!\n")
	return nil
}

func deleteBootstrapCluster(runPlural ActionFunc) {
	if err := ExecuteSteps([]*Step{{
		Name:    "Destroy local cluster",
		Args:    []string{"plural", "--bootstrap", "bootstrap", "cluster", "delete", "bootstrap"},
		Execute: runPlural,
	}}); err != nil {
		utils.Error("%s", err)
	}
}
