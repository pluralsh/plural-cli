package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/imdario/mergo"
	metal "github.com/packethost/packngo"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/template"
	"github.com/pluralsh/plural/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"
	"sigs.k8s.io/yaml"
)

type EQUINIXProvider struct {
	Clust  string `survey:"cluster"`
	Proj   string
	bucket string
	Metro  string `survey:"metro"`
	ctx    map[string]interface{}
}

var equinixSurvey = []*survey.Question{
	{
		Name:     "cluster",
		Prompt:   &survey.Input{Message: "Enter the name of your cluster:"},
		Validate: utils.ValidateAlphaNumeric,
	},
	{
		Name:     "metro",
		Prompt:   &survey.Input{Message: "What metro will you deploy to?", Default: "sv"},
		Validate: survey.Required,
	},
	{
		Name:     "project",
		Prompt:   &survey.Input{Message: "Enter the name of the project you want to use:"},
		Validate: survey.Required,
	},
}

func mkEquinix(conf config.Config) (provider *EQUINIXProvider, err error) {
	var resp struct {
		Clust   string
		Metro   string
		Project string
	}
	if err := survey.Ask(equinixSurvey, &resp); err != nil {
		return nil, err
	}

	projectID, err := getProjectIDFromName(resp.Project)
	if err != nil {
		return nil, utils.ErrorWrap(err, "Failed to get metal project ID (is your metal cli configured?)")
	}

	provider = &EQUINIXProvider{
		resp.Clust,
		projectID,
		"",
		resp.Metro,
		map[string]interface{}{},
	}

	projectManifest := manifest.ProjectManifest{
		Cluster:  provider.Cluster(),
		Project:  provider.Project(),
		Provider: EQUINIX,
		Region:   provider.Region(),
		Context:  provider.Context(),
		Owner:    &manifest.Owner{Email: conf.Email, Endpoint: conf.Endpoint},
	}

	if err := projectManifest.Configure(); err != nil {
		return nil, err
	}

	provider.bucket = projectManifest.Bucket
	return provider, nil
}

func equinixFromManifest(man *manifest.Manifest) (*EQUINIXProvider, error) {
	return &EQUINIXProvider{man.Cluster, man.Project, man.Bucket, man.Region, man.Context}, nil
}

func (equinix *EQUINIXProvider) CreateBackend(prefix string, ctx map[string]interface{}) (string, error) {

	ctx["Region"] = equinix.Region()
	ctx["Bucket"] = equinix.Bucket()
	ctx["Prefix"] = prefix
	ctx["ClusterCreated"] = false
	ctx["__CLUSTER__"] = equinix.Cluster()
	if cluster, ok := ctx["cluster"]; ok {
		ctx["Cluster"] = cluster
		ctx["ClusterCreated"] = true
	} else {
		ctx["Cluster"] = fmt.Sprintf(`"%s"`, equinix.Cluster())
	}

	utils.WriteFile(filepath.Join(equinix.Bucket(), ".gitignore"), []byte("!/*"))
	utils.WriteFile(filepath.Join(equinix.Bucket(), ".gitattributes"), []byte("/* filter=plural-crypt diff=plural-crypt"))

	return template.RenderString(equinixBackendTemplate, ctx)
}

func (equinix *EQUINIXProvider) KubeConfig() error {
	// TODO: deal with current configured KUBECONFIG
	// TODO: deal with KUBECONFIG env var if it is set, as then the output KUBECONFIG file will be used
	if utils.InKubernetes() {
		return nil
	}

	usr, _ := user.Current()

	input, err := ioutil.ReadFile(filepath.Join(usr.HomeDir, ".kube/config"))
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(usr.HomeDir, ".kube/config-plural-backup"), input, 0644)
	if err != nil {
		return err
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	kubeConfigFiles := []string{
		filepath.Join(usr.HomeDir, ".kube/config-plural-backup"),
		// TODO: make file path more robust
		filepath.Join(currentDir, "../bootstrap/terraform/kube_config_cluster.yaml"),
	}
	kubeconfigs := []*clientcmdapi.Config{}

	for _, filename := range kubeConfigFiles {
		if len(filename) == 0 {
			// no work to do
			continue
		}

		config, err := clientcmd.LoadFromFile(filename)

		if err != nil {
			return err
		}

		kubeconfigs = append(kubeconfigs, config)
	}

	// first merge all of our maps
	mapConfig := clientcmdapi.NewConfig()

	for _, kubeconfig := range kubeconfigs {
		mergo.Merge(mapConfig, kubeconfig, mergo.WithOverride)
	}

	// merge all of the struct values
	nonMapConfig := clientcmdapi.NewConfig()
	for i := range kubeconfigs {
		kubeconfig := kubeconfigs[i]
		mergo.Merge(nonMapConfig, kubeconfig, mergo.WithOverride)
	}

	// since values are overwritten, but maps values are not, we can merge the non-map config on top of the map config and
	// get the values we expect.
	config := clientcmdapi.NewConfig()
	mergo.Merge(config, mapConfig, mergo.WithOverride)
	mergo.Merge(config, nonMapConfig, mergo.WithOverride)

	json, err := runtime.Encode(clientcmdlatest.Codec, config)
	if err != nil {
		return err
	}
	output, err := yaml.JSONToYAML(json)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(usr.HomeDir, ".kube/config"), output, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (equinix *EQUINIXProvider) Install() (err error) {
	if exists, _ := utils.Which("metal"); exists {
		utils.Success("metal cli already installed!\n")
		return
	}

	fmt.Println("Equinix Metal requires you to manually pkg install the metal cli")

	fmt.Println("Visit https://github.com/equinix/metal-cli#installation to install")
	return
}

func (equinix *EQUINIXProvider) Name() string {
	return EQUINIX
}

func (equinix *EQUINIXProvider) Cluster() string {
	return equinix.Clust
}

func (equinix *EQUINIXProvider) Project() string {
	return equinix.Proj
}

func (equinix *EQUINIXProvider) Bucket() string {
	return equinix.bucket
}

func (equinix *EQUINIXProvider) Region() string {
	return equinix.Metro
}

func (equinix *EQUINIXProvider) Context() map[string]interface{} {
	return map[string]interface{}{}
}

func (prov *EQUINIXProvider) Decommision(node *v1.Node) error {
	// TODO: Figure out how to get and store API token
	client, err := metal.NewClient()

	if err != nil {
		return utils.ErrorWrap(err, "Failed to create Equinix Metal client")
	}

	deviceID := strings.Replace(node.Spec.ProviderID, "equinixmetal://", "", -1)

	_, err = client.Devices.Delete(deviceID, false)

	if err != nil {
		return utils.ErrorWrap(err, "failed to terminate instance")
	}

	return nil
}

func getProjectIDFromName(projectName string) (string, error) {
	cmd := exec.Command("metal", "project", "get", "-o", "json")
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(out)
		return "", err
	}

	var projectID string
	var res []struct {
		Name string
		Id   string
	}
	json.Unmarshal(out, &res)

	for _, project := range res {
		if project.Name == projectName {
			projectID = project.Id
		}
	}
	if projectID == "" {
		return "", fmt.Errorf("Project with name %s not found", projectName)
	}

	return projectID, nil
}
