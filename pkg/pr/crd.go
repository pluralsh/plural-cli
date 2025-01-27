package pr

import (
	"fmt"
	"os"
	"strings"

	"github.com/samber/lo"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/bundle"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/utils"

	"github.com/pluralsh/console/go/controller/api/v1alpha1"
	"github.com/pluralsh/polly/algorithms"
	"sigs.k8s.io/yaml"
)

func BuildCRD(path, contextFile string) (*PrTemplate, error) {
	var prAutomationOk bool
	pr := &v1alpha1.PrAutomation{}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	yamlDocs := strings.Split(string(data), "---")
	for _, yamlDoc := range yamlDocs {
		// Skip empty documents (may occur if there are extra `---`)
		if len(strings.TrimSpace(yamlDoc)) == 0 {
			continue
		}

		if err := yaml.Unmarshal([]byte(yamlDoc), pr); err != nil {
			return nil, err
		}
		if prAutomationOk = isPrAutomation(pr); prAutomationOk {
			break
		}
	}

	if !prAutomationOk {
		return nil, fmt.Errorf("no pr automation found in %s", path)
	}

	prTemplate := &PrTemplate{
		ApiVersion: pr.APIVersion,
		Kind:       pr.Kind,
		Spec:       PrTemplateSpec{},
	}

	ctx, err := configuration(pr, contextFile)
	if err != nil {
		return nil, err
	}
	prTemplate.Context = ctx

	prTemplate.Spec.Creates = creates(pr)
	prTemplate.Spec.Updates = updates(pr)
	prTemplate.Spec.Deletes = deletes(pr)

	return prTemplate, nil
}

func configuration(pr *v1alpha1.PrAutomation, contextFile string) (map[string]interface{}, error) {
	ctx := map[string]interface{}{}
	if len(pr.Spec.Configuration) == 0 {
		return ctx, nil
	}

	if contextFile != "" {
		err := utils.YamlFile(contextFile, &ctx)
		return ctx, err
	}

	path := manifest.ProjectManifestPath()
	man := manifest.ProjectManifest{}
	if !utils.Exists(path) {
		if err := man.Write(path); err != nil {
			return ctx, fmt.Errorf("error writing manifest: %w", err)
		}
		defer os.Remove(path)
	}

	items := algorithms.Map(pr.Spec.Configuration, func(t v1alpha1.PrAutomationConfiguration) *api.ConfigurationItem {
		ci := &api.ConfigurationItem{
			Name:       t.Name,
			Type:       t.Type.String(),
			Validation: nil,
		}
		if t.Default != nil {
			ci.Default = *t.Default
		}
		if t.Documentation != nil {
			ci.Documentation = *t.Documentation
		}
		if t.Placeholder != nil {
			ci.Placeholder = *t.Placeholder
		}
		if t.Optional != nil {
			ci.Optional = *t.Optional
		}
		if t.Condition != nil {
			condition := &api.Condition{
				Field:     t.Condition.Field,
				Operation: t.Condition.Operation.String(),
			}
			if t.Condition.Value != nil {
				condition.Value = *t.Condition.Value
			}
			ci.Condition = condition
		}
		if t.Validation != nil {
			validation := &api.Validation{}
			if t.Validation.Regex != nil {
				validation.Regex = *t.Validation.Regex
			}
			ci.Validation = validation
		}
		if len(t.Values) > 0 {
			ci.Values = algorithms.Map(t.Values, func(t *string) string {
				return *t
			})
		}
		return ci
	})
	utils.Highlight("Lets' fill out the configuration for this PR automation:\n")
	for _, item := range items {
		if err := bundle.Configure(ctx, item, &manifest.Context{}, ""); err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

func deletes(pr *v1alpha1.PrAutomation) *DeleteSpec {
	if d := pr.Spec.Deletes; d != nil {
		return &DeleteSpec{
			Files:   d.Files,
			Folders: d.Folders,
		}
	}
	return nil
}

func updates(pr *v1alpha1.PrAutomation) *UpdateSpec {
	u := pr.Spec.Updates
	if u == nil {
		return nil
	}
	prUpdates := &UpdateSpec{
		Regexes:           make([]string, 0),
		Files:             make([]string, 0),
		RegexReplacements: make([]RegexReplacement, 0),
		YamlOverlays:      make([]YamlOverlay, 0),
	}
	if u.ReplaceTemplate != nil {
		prUpdates.ReplaceTemplate = *u.ReplaceTemplate
	}
	if u.Yq != nil {
		prUpdates.Yq = *u.Yq
	}
	if u.MatchStrategy != nil {
		prUpdates.MatchStrategy = u.MatchStrategy.String()
	}

	if len(u.Regexes) > 0 {
		prUpdates.Regexes = algorithms.Map(u.Regexes, func(t *string) string {
			return *t
		})
	}
	if len(u.Files) > 0 {
		prUpdates.Files = algorithms.Map(u.Files, func(t *string) string {
			return *t
		})
	}
	if len(u.RegexReplacements) > 0 {
		prUpdates.RegexReplacements = algorithms.Map(u.RegexReplacements, func(t v1alpha1.RegexReplacement) RegexReplacement {
			return RegexReplacement{
				Regex:       t.Regex,
				Replacement: t.Replacement,
				File:        t.File,
				Templated:   lo.FromPtr(t.Templated),
			}
		})
	}
	if len(u.YamlOverlays) > 0 {
		prUpdates.YamlOverlays = algorithms.Map(u.YamlOverlays, func(y v1alpha1.YamlOverlay) YamlOverlay {
			return YamlOverlay{
				File:      y.File,
				Yaml:      y.Yaml,
				Templated: lo.FromPtr(y.Templated),
				ListMerge: toListMerge(y.ListMerge),
			}
		})
	}
	return prUpdates
}

func creates(pr *v1alpha1.PrAutomation) *CreateSpec {
	c := pr.Spec.Creates
	if c == nil {
		return nil
	}
	prCreates := &CreateSpec{
		Templates: make([]*CreateTemplate, 0),
	}
	if c.Git != nil {
		prCreates.ExternalDir = c.Git.Folder
	}
	for _, t := range c.Templates {
		createTemplate := &CreateTemplate{
			Source:      t.Source,
			Destination: t.Destination,
			External:    t.External,
		}
		if t.Context != nil {
			ctx := map[string]interface{}{}
			if err := yaml.Unmarshal(t.Context.Raw, &ctx); err == nil {
				createTemplate.Context = ctx
			} else {
				utils.Error("Failed to unmarshal PrAutomationTemplate context: %v \n", err)
			}

		}
		prCreates.Templates = append(prCreates.Templates, createTemplate)
	}

	return prCreates
}

func isPrAutomation(o client.Object) bool {
	return o.GetObjectKind().GroupVersionKind().Kind == "PrAutomation" && o.GetObjectKind().GroupVersionKind().Group == v1alpha1.GroupName
}
