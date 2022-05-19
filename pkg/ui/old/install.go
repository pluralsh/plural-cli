package uiOld

import (
	"fmt"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/bundle"
	"github.com/pluralsh/plural/pkg/manifest"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/rivo/tview"
	"github.com/urfave/cli"
)

func Install(c *cli.Context, repo, recipeName string) *tview.Form {
	client := api.NewClient()
	recipe, _ := client.GetRecipe(repo, recipeName)

	path := manifest.ContextPath()
	context, err := manifest.ReadContext(path)
	if err != nil {
		context = manifest.NewContext()
	}

	form := tview.NewForm()

	context.AddBundle(repo, recipeName)

	for _, section := range recipe.RecipeSections {

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

			if _, ok := ctx[configItem.Name]; ok { //TODO: add back refresh
				continue
			}

			seen[configItem.Name] = true
			// if err := configure(ctx, configItem, context, section, form); err != nil {
			// 	context.Configuration[section.Repository.Name] = ctx
			// 	context.Write(path)
			// 	return nil
			// }
			switch configItem.Type {
			case bundle.Int:
				form.AddInputField(configItem.Name, configItem.Default, 20, nil, nil)
			// 	var res int
			// 	prompt, opts := intSurvey(def, item, proj)
			// 	survey.AskOne(prompt, &res, opts...)
			// 	ctx[item.Name] = res
			case bundle.Bool:
				form.AddCheckbox(configItem.Name, false, nil)
			case bundle.Domain:
				form.AddInputField(configItem.Name, configItem.Default, 20, nil, nil)
			case bundle.String:
				form.AddInputField(configItem.Name, configItem.Default, 20, nil, nil)
			case bundle.Password:
				form.AddPasswordField(configItem.Name, "", 20, '*', nil)
				// case bundle.Bucket:
				// 	var res string
				// 	def = bundle.PrevDefault(ctx, item, def)
				// 	prompt, opts := bucketSurvey(def, item, proj, context, section)
				// 	survey.AskOne(prompt, &res, opts...)
				// 	if res != def {
				// 		ctx[item.Name] = bundle.BucketName(res, proj)
				// 	} else {
				// 		ctx[item.Name] = res
				// 	}
			}
		}

		context.Configuration[section.Repository.Name] = ctx
	}

	return form
}

// func Install(repo, name string, refresh bool) error {
// 	client := api.NewClient()
// 	recipe, err := client.GetRecipe(repo, name)
// 	if err != nil {
// 		return err
// 	}

// 	path := manifest.ContextPath()
// 	context, err := manifest.ReadContext(path)
// 	if err != nil {
// 		context = manifest.NewContext()
// 	}

// 	context.AddBundle(repo, name)

// 	for _, section := range recipe.RecipeSections {
// 		screen.Clear()
// 		screen.MoveTopLeft()
// 		utils.Highlight(section.Repository.Name)
// 		fmt.Printf(" %s\n", section.Repository.Description)

// 		ctx, ok := context.Configuration[section.Repository.Name]
// 		if !ok {
// 			ctx = map[string]interface{}{}
// 		}

// 		seen := make(map[string]bool)

// 		for _, configItem := range section.Configuration {
// 			if seen[configItem.Name] {
// 				continue
// 			}

// 			if _, ok := ctx[configItem.Name]; ok && !refresh {
// 				continue
// 			}

// 			seen[configItem.Name] = true
// 			if err := configure(ctx, configItem, context, section); err != nil {
// 				context.Configuration[section.Repository.Name] = ctx
// 				context.Write(path)
// 				return err
// 			}
// 		}

// 		context.Configuration[section.Repository.Name] = ctx
// 	}

// 	err = context.Write(path)
// 	if err != nil {
// 		return err
// 	}

// 	if err := performTests(context, recipe); err != nil {
// 		return err
// 	}

// 	err = client.InstallRecipe(recipe.Id)
// 	if err != nil {
// 		return err
// 	}

// 	if recipe.OidcSettings == nil {
// 		return nil
// 	}

// 	confirm := false
// 	if err := configureOidc(repo, client, recipe, context.Configuration[repo], &confirm); err != nil {
// 		return err
// 	}

// 	for _, r := range recipe.RecipeDependencies {
// 		repo := r.Repository.Name
// 		if err := configureOidc(repo, client, r, context.Configuration[repo], &confirm); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func configure(ctx map[string]interface{}, item *api.ConfigurationItem, context *manifest.Context, section *api.RecipeSection) error {
// 	return nil
// }

// func configure(ctx map[string]interface{}, item *api.ConfigurationItem, context *manifest.Context, section *api.RecipeSection, form *tview.Form) error {
// 	if !bundle.EvaluateCondition(ctx, item.Condition) {
// 		return nil
// 	}

// 	// if item.Type == Function {
// 	// 	res, err := fetchFunction(item)
// 	// 	if err != nil {
// 	// 		return err
// 	// 	}
// 	// 	ctx[item.Name] = res
// 	// 	return nil
// 	// }

// 	proj, err := manifest.FetchProject()
// 	if err != nil {
// 		return err
// 	}

// 	fmt.Println("")
// 	utils.Highlight(item.Name)
// 	fmt.Printf("\n>> %s\n", item.Documentation)
// 	def := bundle.GetDefault(item.Default, item, proj)

// 	switch item.Type {
// 	case bundle.Int:
// 		var res int
// 		prompt, opts := intSurvey(def, item, proj)
// 		survey.AskOne(prompt, &res, opts...)
// 		ctx[item.Name] = res
// 	case bundle.Bool:
// 		res := false
// 		prompt, opts := boolSurvey(def, item, proj)
// 		survey.AskOne(prompt, &res, opts...)
// 		ctx[item.Name] = res
// 	case bundle.Domain:
// 		var res string
// 		def = bundle.PrevDefault(ctx, item, def)
// 		prompt, opts := domainSurvey(def, item, proj)
// 		survey.AskOne(prompt, &res, opts...)
// 		ctx[item.Name] = res
// 	case bundle.String:
// 		var res string
// 		def = bundle.PrevDefault(ctx, item, def)
// 		prompt, opts := stringSurvey(def, item, proj)
// 		survey.AskOne(prompt, &res, opts...)
// 		ctx[item.Name] = res
// 	case bundle.Password:
// 		var res string
// 		def = bundle.PrevDefault(ctx, item, def)
// 		prompt, opts := passwordSurvey(def, item, proj)
// 		survey.AskOne(prompt, &res, opts...)
// 		ctx[item.Name] = res
// 	case bundle.Bucket:
// 		var res string
// 		def = bundle.PrevDefault(ctx, item, def)
// 		prompt, opts := bucketSurvey(def, item, proj, context, section)
// 		survey.AskOne(prompt, &res, opts...)
// 		if res != def {
// 			ctx[item.Name] = bundle.BucketName(res, proj)
// 		} else {
// 			ctx[item.Name] = res
// 		}
// 	case bundle.File:
// 		var res string
// 		prompt, opts := fileSurvey(def, item, proj)
// 		survey.AskOne(prompt, &res, opts...)
// 		path, err := homedir.Expand(res)
// 		if err != nil {
// 			return err
// 		}
// 		contents, err := utils.ReadFile(path)
// 		if err != nil {
// 			return err
// 		}
// 		ctx[item.Name] = contents
// 	}

// 	return nil
// }
