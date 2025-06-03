package tests

import (
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/manifest"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

func Perform(ctx *manifest.Context, test *api.RecipeTest) error {
	utils.Highlight("\nRunning %s test [%s] ==>\n", test.Name, test.Type)
	if test.Type == "GIT" {
		return testGit(ctx, test)
	}
	return nil
}
