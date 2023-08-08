package utils

import (
	"github.com/pluralsh/plural/pkg/utils"
)

func FailedPermission(perm string) {
	utils.Highlight("Required permission %s: ", perm)
	utils.Error("failed\n")
}

func WarnRole(role string) {
	utils.Highlight("Recommended role %s: ", role)
	utils.Warn("missing\n")
}
