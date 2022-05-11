package scm

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural/pkg/utils"
)

var validRepo = survey.ComposeValidators(
	utils.ValidateAlphaNumeric,
	survey.MaxLength(20),
)

func repoName() (name string) {
	prompt := &survey.Input{
		Message: "Choose a memorable repo name:",
	}
	survey.AskOne(prompt, &name, survey.WithValidator(validRepo))
	return
}