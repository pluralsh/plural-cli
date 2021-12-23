package bundle

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
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

	fmt.Println("")
	utils.Highlight(item.Name)
	fmt.Printf("\n>> %s\n", item.Documentation)
	def := genDefault(item.Default, item, proj)

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
		prompt, opts := domainSurvey(def, item, proj)
		survey.AskOne(prompt, &res, opts...)
		ctx[item.Name] = res
	case String:
		var res string
		prompt, opts := stringSurvey(def, item, proj)
		survey.AskOne(prompt, &res, opts...)
		ctx[item.Name] = res
	case Password:
		var res string
		prompt, opts := passwordSurvey(def, item, proj)
		survey.AskOne(prompt, &res, opts...)
		ctx[item.Name] = res
	case Bucket:
		var res string
		prompt, opts := bucketSurvey(def, item, proj)
		survey.AskOne(prompt, &res, opts...)
		if res != def {
			ctx[item.Name] = bucketName(res, proj)
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

func genDefault(def string, item *api.ConfigurationItem, proj *manifest.ProjectManifest) string {
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

	return fmt.Sprintf("%s-%s-%s", proj.BucketPrefix, proj.Cluster, value)
}
