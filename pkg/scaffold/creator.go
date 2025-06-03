package scaffold

import (
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/plural-cli/pkg/utils/pathing"
	"helm.sh/helm/v3/pkg/chartutil"
)

var categories = []string{
	"data",
	"productivity",
	"devops",
	"database",
	"messaging",
	"security",
	"network",
}

var scaffoldSurvey = []*survey.Question{
	{
		Name:     "application",
		Prompt:   &survey.Input{Message: "Enter the name of your application:"},
		Validate: utils.ValidateAlphaNumeric,
	},
	{
		Name:     "publisher",
		Prompt:   &survey.Input{Message: "Enter the name of your publisher:"},
		Validate: survey.Required,
	},
	{
		Name: "category",
		Prompt: &survey.Select{
			Message: "Enter the category for your application:",
			Options: categories,
		},
		Validate: survey.Required,
	},
	{
		Name:     "postgres",
		Prompt:   &survey.Confirm{Message: "Will your application need a postgres database?"},
		Validate: survey.Required,
	},
	{
		Name:     "ingress",
		Prompt:   &survey.Confirm{Message: "Does your application need an ingress?"},
		Validate: survey.Required,
	},
}

func ApplicationScaffold(client api.Client) error {
	input := api.ScaffoldInputs{}
	if err := survey.Ask(scaffoldSurvey, &input); err != nil {
		return err
	}

	scaffolds, err := client.Scaffolds(&input)
	if err != nil {
		return api.GetErrorResponse(err, "Scaffolds")
	}

	app := input.Application
	helmPath := pathing.SanitizeFilepath(filepath.Join(app, "helm"))
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(helmPath, 0755); err != nil {
		return err
	}

	if err := os.Chdir(helmPath); err != nil {
		return err
	}

	if err := createHelm(app); err != nil {
		return err
	}

	if err := os.Chdir(pathing.SanitizeFilepath(filepath.Join(pwd, app))); err != nil {
		return err
	}

	for _, scaffold := range scaffolds {
		if err := utils.WriteFile(scaffold.Path, []byte(scaffold.Content)); err != nil {
			return err
		}
	}

	return nil
}

func createHelm(name string) error {
	chartname := filepath.Base(name)
	_, err := chartutil.Create(chartname, filepath.Dir(name))
	return err
}
