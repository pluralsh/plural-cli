package bundle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils"

	homedir "github.com/mitchellh/go-homedir"
)

func stringValidator(item *api.ConfigurationItem) survey.AskOpt {
	return survey.WithValidator(func(val interface{}) error {
		res, _ := val.(string)
		if item.Validation != nil && item.Validation.Type == "REGEX" {
			valid := item.Validation
			return utils.ValidateRegex(res, valid.Regex, valid.Message)
		}
		return nil
	})
}

func stringSurvey(def string, item *api.ConfigurationItem) (survey.Prompt, []survey.AskOpt) {
	opts := []survey.AskOpt{stringValidator(item)}

	if !item.Optional {
		opts = append(opts, survey.WithValidator(survey.Required))
	}

	return &survey.Input{
		Message: "Enter the value",
		Default: def,
	}, opts
}

func passwordSurvey(item *api.ConfigurationItem) (survey.Prompt, []survey.AskOpt) {
	opts := []survey.AskOpt{stringValidator(item)}

	if !item.Optional {
		opts = append(opts, survey.WithValidator(survey.Required))
	}

	return &survey.Password{Message: "Enter the value"}, opts
}

func boolSurvey() (survey.Prompt, []survey.AskOpt) {
	return &survey.Confirm{Message: "Yes or no?"}, []survey.AskOpt{survey.WithValidator(survey.Required)}
}

func intSurvey(def string) (survey.Prompt, []survey.AskOpt) {
	return &survey.Input{
		Message: "Enter the value",
		Default: def,
	}, []survey.AskOpt{survey.WithValidator(survey.Required)}
}

func domainSurvey(def string, item *api.ConfigurationItem, proj *manifest.ProjectManifest, context *manifest.Context) (survey.Prompt, []survey.AskOpt) {
	opts := []survey.AskOpt{
		survey.WithValidator(func(val interface{}) error {
			res, _ := val.(string)
			if res == "" && item.Optional {
				return nil
			}

			if context.HasDomain(res) {
				return fmt.Errorf("domain %s has already been used elsewhere in this project, please chose another", res)
			}

			if proj.Network != nil && !strings.HasSuffix(res, proj.Network.Subdomain) {
				return fmt.Errorf("domain must end with %s", proj.Network.Subdomain)
			}

			if err := utils.ValidateDns(res); err != nil {
				return err
			}

			return nil
		}),
	}

	msg := fmt.Sprintf("Enter a domain, which must be beneath %s ", proj.Network.Subdomain)
	return &survey.Input{Message: msg, Default: def}, opts
}

func fileSurvey(def string) (survey.Prompt, []survey.AskOpt) {
	return &survey.Input{
		Message: "select a file (use tab to list files in the directory):",
		Default: def,
		Suggest: func(toComplete string) []string {
			path, err := homedir.Expand(toComplete)
			if err != nil {
				path = toComplete
			}
			files, _ := filepath.Glob(cleanPath(path) + "*")
			return files
		},
	}, []survey.AskOpt{survey.WithValidator(survey.Required)}
}

func cleanPath(path string) string {
	if strings.HasSuffix(path, "/") {
		return path
	}

	fi, err := os.Stat(path)
	if err != nil {
		return path
	}

	if fi.Mode().IsDir() {
		return path + string(filepath.Separator)
	}

	return path
}

func bucketSurvey(def string, proj *manifest.ProjectManifest, context *manifest.Context) (survey.Prompt, []survey.AskOpt) {
	prompt := "Enter a globally unique object store bucket name "

	if proj.BucketPrefix != "" {
		prompt = fmt.Sprintf("Enter a globally unique bucket name, will be prefixed with %s-%s-", proj.BucketPrefix, proj.Cluster)
	}

	opts := []survey.AskOpt{
		survey.WithValidator(func(val interface{}) error {
			res, _ := val.(string)
			name := bucketName(res, proj)
			if len(name) > 63 || len(name) < 3 {
				return fmt.Errorf("bucket name must be between 3 and 63 characters long")
			}

			if context.HasBucket(name) {
				return fmt.Errorf("bucket %s has already been used elsewhere in this project, please chose another", name)
			}

			if err := utils.ValidateRegex(name, "[a-z][a-z0-9\\-]+[a-z0-9]", "Name must be a hyphenated alphanumeric string"); err != nil {
				return err
			}

			return nil
		}),
		survey.WithValidator(survey.Required),
	}

	return &survey.Input{
		Message: prompt,
		Default: def,
	}, opts
}
