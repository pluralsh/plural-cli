package provider

import (
	"github.com/AlecAivazis/survey/v2"

	"github.com/pluralsh/plural-cli/pkg/utils"
)

var validCluster = survey.ComposeValidators(
	utils.ValidateAlphaNumeric,
	survey.MaxLength(15),
)
