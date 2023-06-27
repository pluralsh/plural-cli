package plural

import (
	"fmt"
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
	BootstrapCluster                bool   `json:"bootstrapCluster"`
	BootstrapCRDS                   bool   `json:"bootstrapCrds"`
	BootstrapDeployCapiOperator     bool   `json:"bootstrapDeployCapiOperator"`
	BootstrapDeployCapiCluster      bool   `json:"bootstrapDeployCapiCluster"`
	BootstrapCapiClusterReady       bool   `json:"bootstrapCapiClusterReady"`
	BootstrapCapiMpReady            bool   `json:"bootstrapCapiMpReady"`
	TargetClusterNamespace          bool   `json:"targetClusterNamespace"`
	TargetClusterCRDS               bool   `json:"targetClusterCRDS"`
	TargetClusterDeployCapiOperator bool   `json:"targetClusterDeployCapiOperator"`
	TargetClusterMoveCluster        bool   `json:"targetClusterMoveCluster"`
	Error                           string `json:"error"`
}

func getStatus() (*ClusterAPIStatus, error) {
	repoRoot, err := git.Root()
	if err != nil {
		return nil, err
	}
	path := pathing.SanitizeFilepath(filepath.Join(repoRoot, "clusterapi.yaml"))
	if !utils.Exists(path) {
		return &ClusterAPIStatus{}, nil
	}
	content, err := utils.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var status *ClusterAPIStatus

	if err := yaml.Unmarshal([]byte(content), &status); err != nil {
		return nil, err
	}
	return status, nil
}

func (c *ClusterAPIStatus) Marshal() ([]byte, error) {
	return yaml.Marshal(&c)
}

func (c *ClusterAPIStatus) isReady(step *Step) bool {
	repoRoot, err := git.Root()
	if err != nil {
		return false
	}
	content, err := utils.ReadFile(pathing.SanitizeFilepath(filepath.Join(repoRoot, "clusterapi.yaml")))
	if err != nil {
		return false
	}

	var status *ClusterAPIStatus

	if err := yaml.Unmarshal([]byte(content), &status); err != nil {
		return false
	}
	ps := reflect.ValueOf(status).Elem()
	field := ps.FieldByName(step.SuccessStatusName)
	if field.IsValid() && field.CanSet() {
		return field.Bool()
	}

	return false
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

func clusterAPIDeploySteps() []*Step {
	pm, _ := manifest.FetchProject()
	root, _ := git.Root()
	sanitizedPath := pathing.SanitizeFilepath(filepath.Join(root, "bootstrap"))

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
			TargetPath:        sanitizedPath,
		},
		{
			Name:              "bootstrap crds",
			Args:              []string{"plural", "--bootstrap", "wkspace", "crds", "bootstrap"},
			SuccessStatusName: "BootstrapCRDS",
			Execute:           RunPlural,
			TargetPath:        sanitizedPath,
		},
		{
			Name:              "install capi operators",
			Args:              append([]string{"plural", "--bootstrap", "wkspace", "helm", "bootstrap", "--skip", "cluster-api-cluster"}, providerBootstrapFlags...),
			SuccessStatusName: "BootstrapDeployCapiOperator",
			Execute:           RunPlural,
			TargetPath:        sanitizedPath,
		},
		{
			Name:              "deploy cluster",
			Args:              append([]string{"plural", "--bootstrap", "wkspace", "helm", "bootstrap"}, providerBootstrapFlags...),
			SuccessStatusName: "BootstrapDeployCapiCluster",
			Execute:           RunPlural,
			TargetPath:        sanitizedPath,
		},
		{
			Name:              "wait-for-cluster",
			Args:              []string{"plural", "--bootstrap", "clusters", "wait", "bootstrap", pm.Cluster},
			SuccessStatusName: "BootstrapCapiClusterReady",
			Execute:           RunPlural,
			TargetPath:        sanitizedPath,
		},
		{
			Name:              "wait-for-machines-running",
			Args:              []string{"plural", "--bootstrap", "clusters", "mpwait", "bootstrap", pm.Cluster},
			SuccessStatusName: "BootstrapCapiMpReady",
			Execute:           RunPlural,
			TargetPath:        sanitizedPath,
		},
		{
			Name:       "init kubeconfig for target cluster",
			Args:       []string{"plural", "wkspace", "kube-init"},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:              "create-bootstrap-namespace-workload-cluster",
			Args:              []string{"plural", "bootstrap", "namespace", "create", "bootstrap"},
			Execute:           RunPlural,
			TargetPath:        sanitizedPath,
			SuccessStatusName: "TargetClusterNamespace",
		},

		{
			Name:              "install CRDs on target cluster",
			Args:              []string{"plural", "wkspace", "crds", "bootstrap"},
			Execute:           RunPlural,
			TargetPath:        sanitizedPath,
			SuccessStatusName: "TargetClusterCRDS",
		},
		{
			Name:              "clusterctl-init-workload",
			Args:              append([]string{"plural", "wkspace", "helm", "bootstrap", "--skip", "cluster-api-cluster"}, providerBootstrapFlags...),
			Execute:           RunPlural,
			TargetPath:        sanitizedPath,
			SuccessStatusName: "TargetClusterDeployCapiOperator",
		},
		{
			Name:              "clusterctl-move",
			Args:              []string{"plural", "bootstrap", "cluster", "move", "--kubeconfig-context", "kind-bootstrap", "--to-kubeconfig", pathing.SanitizeFilepath(filepath.Join(homedir, ".kube", "config"))},
			Execute:           RunPlural,
			TargetPath:        sanitizedPath,
			SuccessStatusName: "TargetClusterMoveCluster",
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
			Execute:    RunTerraform,
			TargetPath: filepath.Join(sanitizedPath, "terraform"),
		},
		{
			Name:       "terraform apply",
			Args:       []string{"apply", "-auto-approve"},
			Execute:    RunTerraform,
			TargetPath: filepath.Join(sanitizedPath, "terraform"),
		},
		{
			Name:       "terraform output",
			Args:       []string{"plural", "output", "terraform", "bootstrap"},
			Execute:    RunPlural,
			TargetPath: filepath.Join(sanitizedPath, "terraform"),
		},
		{
			Name:       "kube init",
			Args:       []string{"plural", "wkspace", "kube-init"},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "crds bootstrap",
			Args:       []string{"plural", "wkspace", "crds", "bootstrap"},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
		{
			Name:       "helm bootstrap",
			Args:       []string{"plural", "wkspace", "helm", "bootstrap", "--skip", "cluster-api-cluster"},
			Execute:    RunPlural,
			TargetPath: sanitizedPath,
		},
	}

}

func ExecuteClusterAPI(path string) error {
	status, err := getStatus()
	if err != nil {
		return err
	}
	for _, step := range clusterAPIDeploySteps() {
		utils.Highlight("%s \n", step.Name)
		err := os.Chdir(step.TargetPath)
		if err != nil {
			return err
		}

		if status.isReady(step) {
			fmt.Println("ready")
			continue
		}
		err = step.Execute(step.Args)
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
		err = os.Chdir(path)
		if err != nil {
			return err
		}
	}
	return nil
}
