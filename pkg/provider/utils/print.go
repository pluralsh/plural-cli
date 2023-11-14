package utils

import (
	"github.com/pluralsh/plural-cli/pkg/utils"
)

func FailedPermission(perm string) {
	utils.Highlight("\nRequired permission %s: ", perm)
	utils.Error("failed\n")
}
