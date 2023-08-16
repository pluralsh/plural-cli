package bootstrap

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/kubernetes"

	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils"
)

// getBootstrapSteps returns list of steps to run during cluster bootstrap.
func getBootstrapSteps(runPlural ActionFunc) ([]*Step, error) {
	projectManifest, err := manifest.FetchProject()
	if err != nil {
		return nil, err
	}

	kubeconfigPath, err := getKubeconfigPath()
	if err != nil {
		return nil, err
	}

	flags := getBootstrapFlags(projectManifest.Provider)

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
			Name:    "Deploy cluster",
			Args:    append([]string{"plural", "--bootstrap", "wkspace", "helm", "bootstrap"}, flags...),
			Execute: runPlural,
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
			Name:    "Destroy local cluster",
			Args:    []string{"plural", "--bootstrap", "bootstrap", "cluster", "delete", "bootstrap"},
			Execute: runPlural,
		},
	}...)
	return steps, nil
}

// BootstrapCluster bootstraps cluster with Cluster API.
func BootstrapCluster(runPlural ActionFunc) error {
	utils.Highlight("Bootstrapping cluster with Cluster API...\n")

	steps, err := getBootstrapSteps(runPlural)
	if err != nil {
		return err
	}

	err = ExecuteSteps(steps)
	if err != nil {
		return err
	}

	utils.Success("Cluster bootstrapped successfully!\n")
	return nil
}