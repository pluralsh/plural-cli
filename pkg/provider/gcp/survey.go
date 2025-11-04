package gcp

import (
	"github.com/AlecAivazis/survey/v2"

	"github.com/pluralsh/plural-cli/pkg/provider/validators"
)

const (
	defaultRegion = "us-east1"
)

// surveyInputProvider is a survey.Ask implementation that reads values from the user.
type surveyInputProvider struct {
	SurveyCluster string `survey:"cluster"`
	SurveyProject string `survey:"project"`
	SurveyRegion  string `survey:"region"`
}

func (in *surveyInputProvider) Cluster() string {
	return in.SurveyCluster
}

func (in *surveyInputProvider) ask(defaultCluster string) error {
	projects, err := Projects()
	if err != nil {
		return err
	}

	err = survey.Ask([]*survey.Question{
		{
			Name:     "cluster",
			Prompt:   &survey.Input{Message: "Enter the name of your cluster", Default: defaultCluster},
			Validate: validators.Cluster(),
		},
		{
			Name:     "project",
			Prompt:   &survey.Select{Message: "Select the GCP project ID:", Options: projects},
			Validate: survey.Required,
		},
	}, in)
	if err != nil {
		return err
	}

	return survey.Ask([]*survey.Question{
		{
			Name: "region",
			Prompt: &survey.Select{
				Message: "What region will you deploy to?",
				Default: defaultRegion,
				Options: Regions(in.Project()),
			},
			Validate: survey.Required,
		},
	}, in)
}

func (in *surveyInputProvider) Project() string {
	return in.SurveyProject
}

func (in *surveyInputProvider) Region() string {
	return in.SurveyRegion
}

func NewSurvey(defaultCluster string) (InputProvider, error) {
	result := new(surveyInputProvider)

	if err := result.ask(defaultCluster); err != nil {
		return nil, err
	}

	return result, nil
}
