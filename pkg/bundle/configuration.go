package bundle

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils"
	homedir "github.com/mitchellh/go-homedir"
)

func evaluateCondition(ctx map[string]interface{}, cond *api.Condition) bool {
	if cond == nil {
		return true
	}

	switch cond.Operation {
	case "NOT":
		val, ok := ctx[cond.Field]
		if !ok { return true }
		booled, ok := val.(bool)
		return ok && !booled
	case "PREFIX":
		val, _ := ctx[cond.Field]
		return strings.HasPrefix(val.(string), cond.Value)
	case "SUFFIX":
		val, _ := ctx[cond.Field]
		return strings.HasSuffix(val.(string), cond.Value)
	}

	return true
}

func configure(ctx map[string]interface{}, item *api.ConfigurationItem) error {
	if !evaluateCondition(ctx, item.Condition) {
		return nil
	}

	if item.Type == Function {
		res, err := fetchFunction(item)
		if err != nil {
			return err
		}
		ctx[item.Name] = res
		return nil
	}

	proj, err := manifest.FetchProject()
	if err != nil {
		return err
	}

	res, def, err := fetchResult(ctx, item, proj)
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

		if err := utils.ValidateDns(res); err != nil {
			return err
		}

		ctx[item.Name] = res
	case String:
		if item.Validation != nil && item.Validation.Type == "REGEX" {
			valid := item.Validation
			if err := utils.ValidateRegex(res, valid.Regex, valid.Message); err != nil {
				return err
			}
		}

		ctx[item.Name] = res
	case Bucket:
		if res != def {
			if err := utils.ValidateRegex(res, "[a-z][a-z0-9\\-]+", "Name must be a hyphenated alphanumeric string"); err != nil {
				return err
			}

			ctx[item.Name] = bucketName(res, proj)
		} else {
			ctx[item.Name] = res
		}
	case File:
		path, err := homedir.Expand(res)
		if err != nil {
			return err
		}
		contents, err := utils.ReadFile(path)
		if err != nil {
			return err
		}
		ctx[item.Name] = contents
	}

	return nil
}

func fetchResult(ctx map[string]interface{}, item *api.ConfigurationItem, proj *manifest.ProjectManifest) (string, string, error) {
	utils.Highlight(item.Name)
	fmt.Printf("\n>> %s\n", item.Documentation)
	prompt := itemPrompt(item, proj)

	def := genDefault(item.Default, item, proj)
	prev, ok := ctx[item.Name]
	if ok {
		def = utils.ToString(prev)
	}

	if def != "" {
		res, err := utils.ReadLineDefault(prompt, def) 
		return res, def, err
	}

	res, err := utils.ReadLine(prompt)
	return res, def, err
}

func genDefault(def string, item *api.ConfigurationItem, proj *manifest.ProjectManifest) string {
	if def == "" {
		return def
	}

	if item.Type != Bucket {
		return def
	}

	return bucketName(def, proj)
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
	case Bucket:
		if proj.BucketPrefix == "" {
			return "Enter a globally unique object store bucket name "
		}

		return fmt.Sprintf("Enter a globally unique bucket name, will be formatted as %s-%s-<your-input>", proj.BucketPrefix, proj.Cluster)
	case String:
		// default
	}

	return "Enter the value "
}

func bucketName(value string, proj *manifest.ProjectManifest) string {
	if proj.BucketPrefix == "" {
		return value
	}

	return fmt.Sprintf("%s-%s-%s", proj.BucketPrefix, proj.Cluster, value)
}
