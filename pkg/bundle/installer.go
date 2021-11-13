package bundle

import (
	"fmt"

	"github.com/inancgumus/screen"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils"
)

func Install(repo, name string) error {
	client := api.NewClient()
	recipe, err := client.GetRecipe(repo, name)
	if err != nil {
		return err
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

		for _, item := range section.RecipeItems {
			for _, configItem := range item.Configuration {
				if seen[configItem.Name] {
					continue
				}

				seen[configItem.Name] = true
				if err := configure(ctx, configItem); err != nil {
					// write current progress to context then return
					context.Configuration[section.Repository.Name] = ctx
					if err := context.Write(path); err != nil {
						return err
					}

					return err
				}
			}
		}

		context.Configuration[section.Repository.Name] = ctx
	}

	err = context.Write(path)
	if err != nil {
		return err
	}

	err = client.InstallRecipe(recipe.Id)
	if err != nil {
		return err
	}

	return configureOidc(repo, client, recipe, context.Configuration[repo])
}

func getName(item *api.RecipeItem) string {
	if item.Terraform != nil {
		return item.Terraform.Name
	}

	if item.Chart != nil {
		return item.Chart.Name
	}

	return ""
}

func getType(item *api.RecipeItem) string {
	if item.Terraform != nil {
		return "terraform"
	}

	if item.Chart != nil {
		return "helm"
	}

	return ""
}

func getDescription(item *api.RecipeItem) string {
	if item.Terraform != nil {
		return item.Terraform.Description
	}

	if item.Chart != nil {
		return item.Chart.Description
	}

	return ""
}
