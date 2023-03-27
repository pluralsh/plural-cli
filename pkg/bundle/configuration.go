package bundle

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mitchellh/go-homedir"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils"
)

type OperationType struct {
	Operation string
	Type      string
}

func evaluateCondition(ctx map[string]interface{}, cond *api.Condition, valueType string) bool {
	if cond == nil {
		return true
	}

	val, ok := ctx[cond.Field]
	if !ok {
		return cond.Operation == "NOT"
	}

	condValue := cond.Value
	operationType := OperationType{
		Operation: cond.Operation,
		Type:      valueType,
	}

	if operationType.Operation == "NOT" {
		booled, ok := val.(bool)
		return ok && !booled // it must be a false boolean value
	}

	if operationType.Operation == "PREFIX" {
		return strings.HasPrefix(val.(string), condValue)
	}

	if operationType.Operation == "SUFFIX" {
		return strings.HasSuffix(val.(string), condValue)
	}

	switch operationType {
	case OperationType{Operation: "EQ", Type: String}:
		return val == condValue
	case OperationType{Operation: "EQ", Type: Int}:
		intCondValue, err := strconv.Atoi(condValue)
		if err != nil {
			return false
		}
		return val.(int) == intCondValue
	case OperationType{Operation: "GT", Type: Int}:
		intCondValue, err := strconv.Atoi(condValue)
		if err != nil {
			return false
		}
		return val.(int) > intCondValue
	case OperationType{Operation: "GTE", Type: Int}:
		intCondValue, err := strconv.Atoi(condValue)
		if err != nil {
			return false
		}
		return val.(int) >= intCondValue
	case OperationType{Operation: "LT", Type: Int}:
		intCondValue, err := strconv.Atoi(condValue)
		if err != nil {
			return false
		}
		return val.(int) < intCondValue
	case OperationType{Operation: "LTE", Type: Int}:
		intCondValue, err := strconv.Atoi(condValue)
		if err != nil {
			return false
		}
		return val.(int) <= intCondValue
	}

	return true
}

func Configure(ctx map[string]interface{}, item *api.ConfigurationItem, context *manifest.Context, repo string) (err error) {
	if !evaluateCondition(ctx, item.Condition, item.Type) {
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
		if value := getEnvVar(repo, item.Name); value != "" {
			res, err = strconv.Atoi(value)
		} else {
			prompt, opts := intSurvey(def)
			err = survey.AskOne(prompt, &res, opts...)
		}
		ctx[item.Name] = res
	case Bool:
		res := false
		if value := getEnvVar(repo, item.Name); value != "" {
			res, err = strconv.ParseBool(value)
		} else {
			prompt, opts := boolSurvey()
			err = survey.AskOne(prompt, &res, opts...)
		}
		ctx[item.Name] = res
	case Domain:
		var res string
		if value := getEnvVar(repo, item.Name); value != "" {
			res = value
		} else {
			def = prevDefault(ctx, item, def)
			prompt, opts := domainSurvey(def, item, proj, context)
			err = survey.AskOne(prompt, &res, opts...)
		}
		ctx[item.Name] = res
		if res != "" {
			context.AddDomain(res)
		}
	case String:
		var res string
		if value := getEnvVar(repo, item.Name); value != "" {
			res = value
		} else {
			def = prevDefault(ctx, item, def)
			prompt, opts := stringSurvey(def, item)
			err = survey.AskOne(prompt, &res, opts...)
		}
		ctx[item.Name] = res
	case Password:
		var res string
		if value := getEnvVar(repo, item.Name); value != "" {
			res = value
		} else {
			prompt, opts := passwordSurvey(item)
			err = survey.AskOne(prompt, &res, opts...)
		}
		ctx[item.Name] = res
	case Bucket:
		var res string
		if value := getEnvVar(repo, item.Name); value != "" {
			ctx[item.Name] = bucketName(value, proj)
		} else {
			def = prevDefault(ctx, item, def)
			prompt, opts := bucketSurvey(def, proj, context)
			err = survey.AskOne(prompt, &res, opts...)
			if res != def {
				ctx[item.Name] = bucketName(res, proj)
			} else {
				ctx[item.Name] = res
			}
		}

		context.AddBucket(ctx[item.Name].(string))
	case File:
		var res string
		if value := getEnvVar(repo, item.Name); value != "" {
			res = value
		} else {
			prompt, opts := fileSurvey(def, item)
			if err := survey.AskOne(prompt, &res, opts...); err != nil {
				return err
			}
		}
		if res == "" {
			return
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

func getEnvVar(repo, itemName string) string {
	variableName := fmt.Sprintf("PLURAL_%s_%s", strings.ToUpper(repo), strings.ToUpper(itemName))
	return os.Getenv(variableName)
}
