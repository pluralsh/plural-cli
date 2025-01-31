package pr

import (
	"os"
	"strings"

	console "github.com/pluralsh/console/go/client"
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
	Deletes *DeleteSpec `json:"deletes"`
}

type UpdateSpec struct {
	Regexes           []string           `json:"regexes"`
	Files             []string           `json:"files"`
	ReplaceTemplate   string             `json:"replace_template"`
	Yq                string             `json:"yq"`
	MatchStrategy     string             `json:"match_strategy"`
	RegexReplacements []RegexReplacement `json:"regex_replacements"`
	YamlOverlays      []YamlOverlay      `json:"yaml_overlays"`
}

type ListMerge string

func toListMerge(listMerge *console.ListMerge) ListMerge {
	// default to overwrite
	if listMerge == nil {
		return ListMergeOverwrite
	}

	switch strings.ToUpper(string(*listMerge)) {
	case string(console.ListMergeOverwrite):
		return ListMergeOverwrite
	case string(console.ListMergeAppend):
		return ListMergeAppend
	}

	return ListMergeOverwrite
}

const (
	ListMergeAppend    = "APPEND"
	ListMergeOverwrite = "OVERWRITE"
)

type YamlOverlay struct {
	File      string    `json:"file"`
	Yaml      string    `json:"yaml"`
	ListMerge ListMerge `json:"list_merge"`
	Templated bool      `json:"templated"`
}

type CreateSpec struct {
	ExternalDir string
	Templates   []*CreateTemplate `json:"templates"`
}

type DeleteSpec struct {
	Files   []string `json:"files"`
	Folders []string `json:"folders"`
}

type CreateTemplate struct {
	Source      string                 `json:"source"`
	Destination string                 `json:"destination"`
	External    bool                   `json:"external"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Condition   string `json:"condition"`
}

type RegexReplacement struct {
	Regex       string `json:"regex"`
	Replacement string `json:"replacement"`
	File        string `json:"file"`
	Templated   bool   `json:"templated"`
}

type PrContracts struct {
	ApiVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   map[string]interface{} `json:"metadata"`
	Context    map[string]interface{} `json:"context"`
	Spec       PrContractsSpec        `json:"spec"`
}

type PrContractsSpec struct {
	Templates   *TemplateCopy        `json:"templates"`
	Workdir     string               `json:"workdir,omitempty"`
	Automations []AutomationContract `json:"automations"`
}

type TemplateCopy struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type AutomationContract struct {
	File        string `json:"file"`
	ExternalDir string `json:"externalDir,omitempty"`
	Context     string `json:"context"`
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

func BuildContracts(path string) (*PrContracts, error) {
	pr := &PrContracts{}
	data, err := os.ReadFile(path)
	if err != nil {
		return pr, err
	}

	if err := yaml.Unmarshal(data, pr); err != nil {
		return pr, err
	}

	return pr, err
}
