package bundle

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mitchellh/go-homedir"
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
		val, ok := ctx[cond.Field]
		if !ok {
			return true
		}
		booled, ok := val.(bool)
		return ok && !booled
	case "PREFIX":
		val := ctx[cond.Field]
		return strings.HasPrefix(val.(string), cond.Value)
	case "SUFFIX":
		val := ctx[cond.Field]
		return strings.HasSuffix(val.(string), cond.Value)
	}

	return true
}

func configure(ctx map[string]interface{}, item *api.ConfigurationItem, context *manifest.Context) (err error) {
	if !evaluateCondition(ctx, item.Condition) {
		return
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
	def := getDefault(item.Default, item, proj)

	switch item.Type {
	case Int:
		var res int
		prompt, opts := intSurvey(def)
		err = survey.AskOne(prompt, &res, opts...)
		ctx[item.Name] = res
	case Bool:
		res := false
		prompt, opts := boolSurvey()
		err = survey.AskOne(prompt, &res, opts...)
		ctx[item.Name] = res
	case Domain:
		var res string
		def = prevDefault(ctx, item, def)
		prompt, opts := domainSurvey(def, item, proj, context)
		err = survey.AskOne(prompt, &res, opts...)
		ctx[item.Name] = res
		if res != "" {
			context.AddDomain(res)
		}
	case String:
		var res string
		def = prevDefault(ctx, item, def)
		prompt, opts := stringSurvey(def, item)
		err = survey.AskOne(prompt, &res, opts...)
		ctx[item.Name] = res
	case Password:
		var res string
		prompt, opts := passwordSurvey(item)
		err = survey.AskOne(prompt, &res, opts...)
		ctx[item.Name] = res
	case Bucket:
		var res string
		def = prevDefault(ctx, item, def)
		prompt, opts := bucketSurvey(def, proj, context)
		err = survey.AskOne(prompt, &res, opts...)
		if res != def {
			ctx[item.Name] = bucketName(res, proj)
		} else {
			ctx[item.Name] = res
		}
		context.AddBucket(ctx[item.Name].(string))
	case File:
		var res string
		prompt, opts := fileSurvey(def)
		if err := survey.AskOne(prompt, &res, opts...); err != nil {
			return err
		}
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

	return
}

func prevDefault(ctx map[string]interface{}, item *api.ConfigurationItem, def string) string {
	if val, ok := ctx[item.Name]; ok {
		if v, ok := val.(string); ok {
			return v
		}
	}

	return def
}

func getDefault(def string, item *api.ConfigurationItem, proj *manifest.ProjectManifest) string {
	if def == "" {
		return def
	}

	if item.Type != Bucket {
		return def
	}

	return bucketName(def, proj)
}

func bucketName(value string, proj *manifest.ProjectManifest) string {
	if proj.BucketPrefix == "" {
		return value
	}

	if strings.HasPrefix(value, proj.BucketPrefix) {
		return value
	}

	return fmt.Sprintf("%s-%s-%s", proj.BucketPrefix, proj.Cluster, value)
}
