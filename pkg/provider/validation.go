package provider

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural/pkg/utils"
)

var validCluster = survey.ComposeValidators(
	utils.ValidateAlphaNumeric,
	survey.MaxLength(15),
)
