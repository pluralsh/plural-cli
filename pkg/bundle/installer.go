package bundle

import (
	"fmt"
	"os"

	"github.com/inancgumus/screen"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/bundle/tests"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils"
)

func Install(repo, name string, refresh bool) error {
	client := api.NewClient()
	recipe, err := client.GetRecipe(repo, name)
	if err != nil {
		return err
	}

	if recipe.Restricted && os.Getenv("CLOUD_SHELL") == "1" {
		return fmt.Errorf("Cannot install this bundle in cloud shell, this is often because it requires a file locally available on your machine like a git ssh key")
	}

	path := manifest.ContextPath()
	context, err := manifest.ReadContext(path)
	if err != nil {
		context = manifest.NewContext()
	}

	context.AddBundle(repo, name)

	for _, section := range recipe.RecipeSections {
		screen.Clear()
		screen.MoveTopLeft()
		utils.Highlight(section.Repository.Name)
		fmt.Printf(" %s\n", section.Repository.Description)

		ctx, ok := context.Configuration[section.Repository.Name]
		if !ok {
			ctx = map[string]interface{}{}
		}

		seen := make(map[string]bool)

		for _, configItem := range section.Configuration {
			if seen[configItem.Name] {
				continue
			}

			if _, ok := ctx[configItem.Name]; ok && !refresh {
				continue
			}

			seen[configItem.Name] = true
			if err := configure(ctx, configItem, context); err != nil {
				context.Configuration[section.Repository.Name] = ctx
				err := context.Write(path)
				if err != nil {
					return err
				}
				return err
			}
		}

		context.Configuration[section.Repository.Name] = ctx
	}

	err = context.Write(path)
	if err != nil {
		return err
	}

	if err := performTests(context, recipe); err != nil {
		return err
	}

	if err := client.InstallRecipe(recipe.Id); err != nil {
		return fmt.Errorf("Install failed, does your plural user have install permissions? error: %w", err)
	}

	if recipe.OidcSettings == nil {
		return nil
	}

	confirm := false
	if err := configureOidc(repo, client, recipe, context.Configuration[repo], &confirm); err != nil {
		return err
	}

	for _, r := range recipe.RecipeDependencies {
		repo := r.Repository.Name
		if err := configureOidc(repo, client, r, context.Configuration[repo], &confirm); err != nil {
			return err
		}
	}

	return nil
}

func performTests(ctx *manifest.Context, recipe *api.Recipe) error {
	if len(recipe.Tests) == 0 {
		return nil
	}

	utils.Highlight("Found %d tests to run...\n", len(recipe.Tests))
	for _, test := range recipe.Tests {
		if err := tests.Perform(ctx, test); err != nil {
			return err
		}
	}

	return nil
}
