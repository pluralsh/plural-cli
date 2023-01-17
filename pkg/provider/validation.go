package provider

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/pluralsh/plural/pkg/utils"
)

var validCluster = survey.ComposeValidators(
	utils.ValidateAlphaNumeric,
	survey.MaxLength(15), //GKE max is 40, EKS max is 100, AKS 53
)

func validClusterName(provider string) survey.Validator {
	var length int
	length = 15
	switch provider {
	case GCP:
		length = 40
	case AWS:
		length = 100
	case AZURE:
		length = 63
	}
	return survey.ComposeValidators(
		utils.ValidateAlphaNumeric,
		survey.MaxLength(length), //GKE max is 40, EKS max is 100, AKS 63
	)
}
