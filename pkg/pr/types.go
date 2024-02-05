package pr

import (
	"os"

	"sigs.k8s.io/yaml"
)

type PrTemplate struct {
	ApiVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   map[string]interface{} `json:"metadata"`
	Context    map[string]interface{} `json:"context"`
	Spec       PrTemplateSpec         `json:"spec"`
}

type PrTemplateSpec struct {
	Updates *UpdateSpec `json:"updates"`
	Creates *CreateSpec `json:"creates"`
}

type UpdateSpec struct {
	Regexes         []string `json:"regexes"`
	Files           []string `json:"files"`
	ReplaceTemplate string   `json:"replace_template"`
	Yq              string   `json:"yq"`
	MatchStrategy   string   `json:"match_strategy"`
}

type CreateSpec struct {
	ExternalDir string
	Templates   []*CreateTemplate `json:"templates"`
}

type CreateTemplate struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	External    bool   `json:"external"`
}

func Build(path string) (*PrTemplate, error) {
	pr := &PrTemplate{}
	data, err := os.ReadFile(path)
	if err != nil {
		return pr, err
	}

	if err := yaml.Unmarshal(data, pr); err != nil {
		return pr, err
	}

	return pr, nil
}
