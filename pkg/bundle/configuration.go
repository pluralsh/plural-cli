package bundle

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils"
)

func EvaluateCondition(ctx map[string]interface{}, cond *api.Condition) bool {
	if cond == nil {
		return true
	}

	switch cond.Operation {
	case "NOT":
		val, ok := ctx[cond.Field]
		if !ok {
			return true
		}
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

func configure(ctx map[string]interface{}, item *api.ConfigurationItem, context *manifest.Context, section *api.RecipeSection) error {
	if !EvaluateCondition(ctx, item.Condition) {
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

	fmt.Println("")
	utils.Highlight(item.Name)
	fmt.Printf("\n>> %s\n", item.Documentation)
	def := GetDefault(item.Default, item, proj)

	switch item.Type {
	case Int:
		var res int
		prompt, opts := intSurvey(def, item, proj)
		survey.AskOne(prompt, &res, opts...)
		ctx[item.Name] = res
	case Bool:
		res := false
		prompt, opts := boolSurvey(def, item, proj)
		survey.AskOne(prompt, &res, opts...)
		ctx[item.Name] = res
	case Domain:
		var res string
		def = PrevDefault(ctx, item, def)
		prompt, opts := domainSurvey(def, item, proj)
		survey.AskOne(prompt, &res, opts...)
		ctx[item.Name] = res
	case String:
		var res string
		def = PrevDefault(ctx, item, def)
		prompt, opts := stringSurvey(def, item, proj)
		survey.AskOne(prompt, &res, opts...)
		ctx[item.Name] = res
	case Password:
		var res string
		def = PrevDefault(ctx, item, def)
		prompt, opts := passwordSurvey(def, item, proj)
		survey.AskOne(prompt, &res, opts...)
		ctx[item.Name] = res
	case Bucket:
		var res string
		def = PrevDefault(ctx, item, def)
		prompt, opts := bucketSurvey(def, item, proj, context, section)
		survey.AskOne(prompt, &res, opts...)
		if res != def {
			ctx[item.Name] = BucketName(res, proj)
		} else {
			ctx[item.Name] = res
		}
	case File:
		var res string
		prompt, opts := fileSurvey(def, item, proj)
		survey.AskOne(prompt, &res, opts...)
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

func PrevDefault(ctx map[string]interface{}, item *api.ConfigurationItem, def string) string {
	if val, ok := ctx[item.Name]; ok {
		if v, ok := val.(string); ok {
			return v
		}
	}

	return def
}

func GetDefault(def string, item *api.ConfigurationItem, proj *manifest.ProjectManifest) string {
	if def == "" {
		return def
	}

	if item.Type != Bucket {
		return def
	}

	return BucketName(def, proj)
}

func BucketName(value string, proj *manifest.ProjectManifest) string {
	if proj.BucketPrefix == "" {
		return value
	}

	if strings.HasPrefix(value, proj.BucketPrefix) {
		return value
	}

	return fmt.Sprintf("%s-%s-%s", proj.BucketPrefix, proj.Cluster, value)
}
