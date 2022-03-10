package provider

import (
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/AlecAivazis/survey/v2"
)

var validCluster = survey.ComposeValidators(
	utils.ValidateAlphaNumeric,
	survey.MaxLength(15),
)