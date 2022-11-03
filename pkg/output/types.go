package output

import (
	"os"

	"github.com/pluralsh/plural/pkg/manifest"
	"gopkg.in/yaml.v2"
)

type Output struct {
	Terraform map[string]interface{}
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
		Kind:       "Output",
		Metadata:   &manifest.Metadata{Name: app},
		Spec:       out,
	}

	io, err := yaml.Marshal(&versioned)
	if err != nil {
		return err
	}

	return os.WriteFile(path, io, 0644)
}

func Read(path string) (out *Output, err error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return
	}

	versioned := &VersionedOutput{Spec: &Output{}}
	err = yaml.Unmarshal(contents, versioned)
	out = versioned.Spec
	return
}
