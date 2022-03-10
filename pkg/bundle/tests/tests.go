package tests

import (
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/manifest"
)

func Perform(ctx *manifest.Context, test *api.RecipeTest) error {
	utils.Highlight("\nRunning %s test [%s] ==>\n", test.Name, test.Type)
	switch test.Type {
	case "GIT":
		return testGit(ctx, test)		
	}
	return nil
}