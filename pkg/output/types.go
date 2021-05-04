package output

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"github.com/pluralsh/plural/pkg/manifest"
)

type Output struct {
	Terraform map[string]string
}

type VersionedOutput struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string
	Metadata   *manifest.Metadata
	Spec       *Output
}

func New() *Output {
	return &Output{}
}

func (out *Output) Save(app, path string) error {
	versioned := &VersionedOutput{
		ApiVersion: "plural.sh/v1alpha1",
		Kind: "Output",
		Metadata: &manifest.Metadata{Name: app},
		Spec: out,
	}

	io, err := yaml.Marshal(&versioned)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, io, 0644)
}

func Read(path string) (out *Output, err error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	versioned := &VersionedOutput{Spec: &Output{}}
	err = yaml.Unmarshal(contents, versioned)
	out = versioned.Spec
	return
}