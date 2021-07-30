package manifest

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
	"github.com/pluralsh/plural/pkg/utils"
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
}

type ProjectManifest struct {
	Cluster  string
	Bucket   string
	Project  string
	Provider string
	Region   string
	Context  map[string]interface{}
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

func (m *ProjectManifest) Write(path string) error {
	versioned := &VersionedProjectManifest{
		ApiVersion: "plural.sh/v1alpha1",
		Kind: "ProjectManifest",
		Metadata: &Metadata{Name: m.Cluster},
		Spec: m,
	}

	io, err := yaml.Marshal(&versioned)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, io, 0644)
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
		Kind: "Manifest",
		Metadata: &Metadata{Name: m.Name},
		Spec: m,
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
