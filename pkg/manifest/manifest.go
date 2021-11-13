package manifest

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/utils"
	"gopkg.in/yaml.v2"
)

type ChartManifest struct {
	Id        string
	Name      string
	VersionId string
	Version   string
}

type TerraformManifest struct {
	Id   string
	Name string
}

type Dependency struct {
	Repo string
}

type Metadata struct {
	Name   string
	Labels map[string]string `yaml:",omitempty"`
}

type Manifest struct {
	Id           string
	Name         string
	Cluster      string
	Project      string
	Bucket       string
	Provider     string
	Region       string
	License      string
	Charts       []*ChartManifest
	Terraform    []*TerraformManifest
	Dependencies []*Dependency
	Context      map[string]interface{}
	Links        *Links `yaml:"links,omitempty"`
}

type Owner struct {
	Email    string
	Endpoint string `yaml:"endpoint,omitempty"`
}

type NetworkConfig struct {
	Subdomain string
	PluralDns bool
}

type ProjectManifest struct {
	Cluster      string
	Bucket       string
	Project      string
	Provider     string
	Region       string
	Owner        *Owner
	Network      *NetworkConfig
	BucketPrefix string `yaml:"bucketPrefix"`
	Context      map[string]interface{}
}

type VersionedManifest struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string
	Metadata   *Metadata
	Spec       *Manifest
}

type VersionedProjectManifest struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string
	Metadata   *Metadata
	Spec       *ProjectManifest
}

func ProjectManifestPath() string {
	root, found := utils.ProjectRoot()
	if !found {
		path, _ := filepath.Abs("workspace.yaml")
		return path
	}

	return filepath.Join(root, "workspace.yaml")
}

func ManifestPath(repo string) (string, error) {
	root, found := utils.ProjectRoot()
	if !found {
		return "", fmt.Errorf("You're not within an installation repo")
	}

	return filepath.Join(root, repo, "manifest.yaml"), nil
}

func (m *ProjectManifest) Write(path string) error {
	versioned := &VersionedProjectManifest{
		ApiVersion: "plural.sh/v1alpha1",
		Kind:       "ProjectManifest",
		Metadata:   &Metadata{Name: m.Cluster},
		Spec:       m,
	}

	io, err := yaml.Marshal(&versioned)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, io, 0644)
}

func FetchProject() (*ProjectManifest, error) {
	path := ProjectManifestPath()
	return ReadProject(path)
}

func ReadProject(path string) (man *ProjectManifest, err error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	versioned := &VersionedProjectManifest{}
	err = yaml.Unmarshal(contents, versioned)
	if err != nil || versioned.Spec == nil {
		man = &ProjectManifest{}
		err = yaml.Unmarshal(contents, man)
		return
	}

	man = versioned.Spec
	return
}

func (m *Manifest) Write(path string) error {
	versioned := &VersionedManifest{
		ApiVersion: "plural.sh/v1alpha1",
		Kind:       "Manifest",
		Metadata:   &Metadata{Name: m.Name},
		Spec:       m,
	}

	io, err := yaml.Marshal(&versioned)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, io, 0644)
}

func Read(path string) (man *Manifest, err error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	versioned := &VersionedManifest{}
	err = yaml.Unmarshal(contents, versioned)
	if err != nil || versioned.Spec == nil {
		man = &Manifest{}
		err = yaml.Unmarshal(contents, man)
		return
	}

	man = versioned.Spec
	return
}

func (man *ProjectManifest) Configure() error {
	utils.Highlight("Let's get some final information about your workspace set up")

	res, _ := utils.ReadAlphaNum("Give us a unique, memorable string to use for bucket naming, eg an abbreviation for your company")
	man.BucketPrefix = res

	if err := man.ConfigureNetwork(); err != nil {
		return err
	}

	return nil
}

func (man *ProjectManifest) ConfigureNetwork() error {
	if man.Network != nil {
		return nil
	}

	utils.Highlight("Ok, let's get your network configuration set up now...\n")
	res, _ := utils.ReadLine("Do you want to use plural's dns provider: [Yn] ")
	pluralDns := res != "n"
	modifier := " (eg something.mydomain.com)"
	if pluralDns {
		modifier = ", must be a subdomain under onplural.sh"
	}
	
	subdomain, _ := utils.ReadLine(fmt.Sprintf("What do you want to use as your subdomain%s: ", modifier))
	if err := utils.ValidateDns(subdomain); err != nil {
		return err
	}

	man.Network = &NetworkConfig{Subdomain: subdomain, PluralDns: pluralDns}

	if pluralDns {
		client := api.NewClient()
		return client.CreateDomain(subdomain)
	}

	return nil
}
