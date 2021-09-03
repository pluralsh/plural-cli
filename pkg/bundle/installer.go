package bundle

import (
	"fmt"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/inancgumus/screen"
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

	for _, section := range recipe.RecipeSections {
		screen.Clear()
		screen.MoveTopLeft()
		utils.Highlight(section.Repository.Name)
		fmt.Printf(" %s\n", section.Repository.Description)

		ctx, ok := context.Configuration[section.Repository.Name]
		if !ok {
			ctx = map[string]interface{}{}
		}

		for _, item := range section.RecipeItems {
			fmt.Printf("\n%s [%s] -- %s\n", getName(item), getType(item), getDescription(item))
			for _, configItem := range item.Configuration {
				if err := configure(ctx, configItem); err != nil {
					return err
				}
			}
		}
		context.Configuration[section.Repository.Name] = ctx
	}

	context.AddBundle(repo, name)
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