package plural

import (
	"os"
	"path/filepath"
	"reflect"

	"github.com/pluralsh/plural/pkg/manifest"

	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"sigs.k8s.io/yaml"
)

type ActionFunc func(arguments []string) error

type Step struct {
	Name              string
	Args              []string
	TargetPath        string
	SuccessStatusName string
	Execute           ActionFunc
}

type ClusterAPIStatus struct {
	BootstrapCluster            bool   `json:"bootstrapCluster"`
	BootstrapCRDS               bool   `json:"bootstrapCrds"`
	BootstrapDeployCapiOperator bool   `json:"bootstrapDeployCapiOperator"`
	BootstrapDeployCapiCluster  bool   `json:"bootstrapDeployCapiCluster"`
	BootstrapCapiClusterReady   bool   `json:"bootstrapCapiClusterReady"`
	BootstrapCapiMpReady        bool   `json:"bootstrapCapiMpReady"`
	Error                       string `json:"error"`
}

func (c *ClusterAPIStatus) Marshal() ([]byte, error) {
	return yaml.Marshal(&c)
}

func (c *ClusterAPIStatus) Save() error {
	data, err := c.Marshal()
	if err != nil {
		return err
	}
	repoRoot, err := git.Root()
	if err != nil {
		return err
	}
	sanitizedPath := pathing.SanitizeFilepath(filepath.Join(repoRoot, "clusterapi.yaml"))
	return os.WriteFile(sanitizedPath, data, 0644)
}

func clusterAPIDeploySteps(path string) []*Step {
	pm, _ := manifest.FetchProject()

	sanitizedPath := pathing.SanitizeFilepath(path)

	homedir, _ := os.UserHomeDir()
	providerBootstrapFlags := []string{}

	switch pm.Provider {
	case "aws":
		providerBootstrapFlags = []string{
			"--set", "cluster-api-provider-aws.cluster-api-provider-aws.bootstrapMode=true",
			"--set", "bootstrap.aws-ebs-csi-driver.enabled=false",
			"--set", "bootstrap.aws-load-balancer-controller.enabled=false",
			"--set", "bootstrap.cluster-autoscaler.enabled=false",
			"--set", "bootstrap.metrics-server.enabled=false",
			"--set", "bootstrap.snapshot-controller.enabled=false",
			"--set", "bootstrap.snapshot-validation-webhook.enabled=false",
			"--set", "bootstrap.tigera-operator.enabled=false",
		}
	case "azure":
		providerBootstrapFlags = []string{}
	case "gcp":
		providerBootstrapFlags = []string{}
	case "google":
		providerBootstrapFlags = []string{}
	}

	return []*Step{
		{
			Name:              "create bootstrap cluster",
			Args:              []string{"plural", "bootstrap", "cluster", "create", "bootstrap", "--skip-if-exists"},
			SuccessStatusName: "BootstrapCluster",
			Execute:           RunPlural,
		},
		{
			Name:              "bootstrap crds",
			Args:              []string{"plural", "--bootstrap", "wkspace", "crds", sanitizedPath},
			SuccessStatusName: "BootstrapCRDS",
			Execute:           RunPlural,
		},
		{
			Name:              "install capi operators",
			Args:              append([]string{"plural", "--bootstrap", "wkspace", "helm", sanitizedPath, "--skip", "cluster-api-cluster"}, providerBootstrapFlags...),
			SuccessStatusName: "BootstrapDeployCapiOperator",
			Execute:           RunPlural,
		},
		{
			Name:              "deploy cluster",
			Args:              append([]string{"plural", "--bootstrap", "wkspace", "helm", sanitizedPath}, providerBootstrapFlags...),
			SuccessStatusName: "BootstrapDeployCapiCluster",
			Execute:           RunPlural,
		},
		{
			Name:              "wait-for-cluster",
			Args:              []string{"plural", "--bootstrap", "clusters", "wait", "bootstrap", pm.Cluster},
			SuccessStatusName: "BootstrapCapiClusterReady",
			Execute:           RunPlural,
		},
		{
			Name:              "wait-for-machines-running",
			Args:              []string{"plural", "--bootstrap", "clusters", "mpwait", "bootstrap", pm.Cluster},
			SuccessStatusName: "BootstrapCapiMpReady",
			Execute:           RunPlural,
		},
		{
			Name:    "init kubeconfig for target cluster",
			Args:    []string{"plural", "wkspace", "kube-init"},
			Execute: RunPlural,
		},
		{
			Name:    "create-bootstrap-namespace-workload-cluster",
			Args:    []string{"plural", "bootstrap", "namespace", "create", "bootstrap"},
			Execute: RunPlural,
		},

		{
			Name:    "crds-bootstrap",
			Args:    []string{"plural", "wkspace", "crds", sanitizedPath},
			Execute: RunPlural,
		},

		{
			Name:    "create-bootstrap-namespace-workload-cluster",
			Args:    []string{"plural", "bootstrap", "namespace", "create", "bootstrap"},
			Execute: RunPlural,
		},
		{
			Name:    "clusterctl-init-workfload",
			Args:    append([]string{"plural", "wkspace", "helm", sanitizedPath, "--skip", "cluster-api-cluster"}, providerBootstrapFlags...),
			Execute: RunPlural,
		},
		{
			Name:    "clusterctl-move",
			Args:    []string{"plural", "bootstrap", "cluster", "move", "--kubeconfig-context", "kind-bootstrap", "--to-kubeconfig", pathing.SanitizeFilepath(filepath.Join(homedir, ".kube", "config"))},
			Execute: RunPlural,
		},
		// { // TODO: re-anable this once we've debugged the move command so it works properly to avoid dangling resources
		// 	Name:    "delete bootstrap cluster",
		// 	Target:  pluralFile(path, "ONCE"),
		// 	Command: "plural",
		// 	Args:    []string{"--bootstrap", "bootstrap", "cluster", "delete", "bootstrap"},
		// 	Sha:     "",
		// },
		{
			Name:       "terraform init",
			Args:       []string{"init", "-upgrade"},
			TargetPath: filepath.Join(path, "terraform"),
			Execute:    RunTerraform,
		},
		{
			Name:       "terraform apply",
			Args:       []string{"apply", "-auto-approve"},
			TargetPath: filepath.Join(path, "terraform"),
			Execute:    RunTerraform,
		},
	}

}

func ExecuteClusterAPI(path, repo string) error {
	err := os.Chdir(path)
	if err != nil {
		return err
	}

	status := &ClusterAPIStatus{}

	for _, step := range clusterAPIDeploySteps(repo) {
		utils.Highlight("%s \n", step.Name)
		err := step.Execute(step.Args)
		if err != nil {
			status.Error = err.Error()
			status.Save()
			return err
		}
		ps := reflect.ValueOf(status).Elem()
		field := ps.FieldByName(step.SuccessStatusName)
		if field.IsValid() && field.CanSet() {
			field.SetBool(true)
		}
		status.Save()
	}
	return nil
}
