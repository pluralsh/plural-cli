package validators

import (
	"github.com/AlecAivazis/survey/v2"

	"github.com/pluralsh/plural-cli/pkg/utils"
)

func Cluster() survey.Validator {
	return survey.ComposeValidators(
		utils.ValidateAlphaNumeric,
		survey.MaxLength(15),
	)
}
