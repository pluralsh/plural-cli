package bundle

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils"
)

func evaluateCondition(ctx map[string]interface{}, cond *api.Condition) bool {
	if cond == nil {
		return true
	}

	switch cond.Operation {
	case "NOT":
		val, _ := ctx[cond.Field]
		return !(val.(bool))
	}

	return true
}

func configure(ctx map[string]interface{}, item *api.ConfigurationItem) error {
	if !evaluateCondition(ctx, item.Condition) {
		return nil
	}

	proj, err := manifest.FetchProject()
	if err != nil {
		return err
	}

	res, err := fetchResult(ctx, item, proj)
	if err != nil {
		return err
	}

	switch item.Type {
	case Int:
		parsed, err := strconv.Atoi(res)
		if err != nil {
			return err
		}
		ctx[item.Name] = parsed
	case Bool:
		parsed, err := strconv.ParseBool(res)
		if err != nil {
			return err
		}
		ctx[item.Name] = parsed
	case Domain:
		if proj.Network != nil && !strings.HasSuffix(res, proj.Network.Subdomain) {
			return fmt.Errorf("Domain must end with %s", proj.Network.Subdomain)
		}
		ctx[item.Name] = res
	case String:
		ctx[item.Name] = res
	}

	return nil
}

func fetchResult(ctx map[string]interface{}, item *api.ConfigurationItem, proj *manifest.ProjectManifest) (string, error) {
	utils.Highlight(item.Name)
	fmt.Printf("\n>> %s\n", item.Documentation)
	prompt := itemPrompt(item, proj)

	def := item.Default
	prev, ok := ctx[item.Name]
	if ok {
		def = utils.ToString(prev)
	}

	if def != "" {
		return utils.ReadLineDefault(prompt, def)
	}

	return utils.ReadLine(prompt)
}

func itemPrompt(item *api.ConfigurationItem, proj *manifest.ProjectManifest) string {
	switch item.Type {
	case Int:
		return "Enter the value (must be an integer) "
	case Bool:
		return "Enter the value (true/false) "
	case Domain:
		if proj.Network != nil {
			return fmt.Sprintf("Enter a domain, which must be beneath %s ", proj.Network.Subdomain)
		}

		return "Enter a domain "
	case String:
		// default
	}

	return "Enter the value "
}
