package api

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

type RecipeInput struct {
	Name         string
	Description  string
	Provider     string
	Sections     []RecipeSectionInput
	Dependencies []DependencyInput
}

type DependencyInput struct {
	Name string
	Repo string
}

type RecipeSectionInput struct {
	Name  string
	Items []RecipeItemInput
}

type RecipeItemInput struct {
	Name          string
	Type          string
	Configuration []ConfigurationItemInput
}

type ConditionInput struct {
	Field     string
	Value     string
	Operation string
}

type ConfigurationItemInput struct {
	Name          string
	Default       string
	Type          string
	Documentation string
	Placeholder   string
	Condition     *ConditionInput
}

type RecipeEdge struct {
	Node *Recipe
}

const createRecipe = `
	mutation CreateRecipe($name: String!, $attributes: RecipeAttributes!) {
		createRecipe(repositoryName: $name, attributes: $attributes) {
			id
		}
	}
`

var getRecipe = fmt.Sprintf(`
query Recipe($repo: String, $name: String) {
	recipe(repo: $repo, name: $name) {
		...RecipeFragment
		recipeSections { ...RecipeSectionFragment }
	}
}
%s
%s
`, RecipeFragment, RecipeSectionFragment)

var listRecipes = fmt.Sprintf(`
query Recipes($repo: String, $provider: Provider) {
	recipes(repositoryName: $repo, provider: $provider, first: 500) {
		edges { node { ...RecipeFragment } }
	}
}
%s
`, RecipeFragment)

const installRecipe = `
mutation Install($id: ID!, $ctx: Map!) {
	installRecipe(recipeId: $id, context: $ctx) {
		id
	}
}
`

func (client *Client) CreateRecipe(repoName string, attrs RecipeInput) (string, error) {
	var resp struct {
		Id string
	}
	req := client.Build(createRecipe)
	req.Var("attributes", attrs)
	req.Var("name", repoName)
	err := client.Run(req, &resp)
	return resp.Id, err
}

func (client *Client) GetRecipe(repo, name string) (recipe *Recipe, err error) {
	var resp struct {
		Recipe *Recipe
	}
	req := client.Build(getRecipe)
	req.Var("repo", repo)
	req.Var("name", name)
	err = client.Run(req, &resp)
	recipe = resp.Recipe
	return
}

func (client *Client) ListRecipes(repo, provider string) (recipes []*Recipe, err error) {
	var resp struct {
		Recipes struct {
			Edges []*RecipeEdge
		}
	}

	req := client.Build(listRecipes)
	req.Var("repo", repo)
	req.Var("provider", provider)
	err = client.Run(req, &resp)
	if err != nil {
		return
	}

	recipes = make([]*Recipe, 0)
	for _, edge := range resp.Recipes.Edges {
		recipes = append(recipes, edge.Node)
	}
	return
}

func (client *Client) InstallRecipe(id string) error {
	var resp struct {
		InstallRecipe []struct {
			Id string
		}
	}

	req := client.Build(installRecipe)
	req.Var("id", id)
	req.Var("ctx", "{}")
	return client.Run(req, &resp)
} 

func ConstructRecipe(marshalled []byte) (RecipeInput, error) {
	var recipe RecipeInput
	err := yaml.Unmarshal(marshalled, &recipe)
	return recipe, err
}
