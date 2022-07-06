package api

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type RecipeInput struct {
	Name         string
	Description  string
	Provider     string
	Restricted   bool
	Tests        []RecipeTestInput `yaml:"tests" json:"tests,omitempty"`
	Sections     []RecipeSectionInput
	Dependencies []DependencyInput
	OidcSettings *OIDCSettings `yaml:"oidcSettings,omitempty"`
}

type DependencyInput struct {
	Name string
	Repo string
}

type RecipeTestInput struct {
	Name    string
	Message string
	Type    string
	Args    []*TestArgInput
}

type TestArgInput struct {
	Name string
	Repo string
	Key  string
}

type RecipeSectionInput struct {
	Name          string
	Items         []RecipeItemInput
	Configuration []ConfigurationItemInput
}

type RecipeItemInput struct {
	Name string
	Type string
}

type ConditionInput struct {
	Field     string
	Value     string
	Operation string
}

type ValidationInput struct {
	Type    string
	Regex   string
	Message string
}

type ConfigurationItemInput struct {
	Name          string
	Default       string
	Type          string
	Documentation string
	Placeholder   string
	Longform      string
	Optional      bool
	FunctionName  string `yaml:"functionName,omitempty" json:"functionName,omitempty"`
	Condition     *ConditionInput
	Validation    *ValidationInput
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
		recipeDependencies { ...RecipeFragment }
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

func (client *Client) CreateRecipe(repoName string, attrs *RecipeInput) (string, error) {
	var resp struct {
		Id string
	}

	if len(attrs.Tests) == 0 {
		attrs.Tests = make([]RecipeTestInput, 0)
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
	if provider != "" {
		req.Var("provider", NormalizeProvider(provider))
	}
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

func ConstructRecipe(marshalled []byte) (recipe RecipeInput, err error) {
	err = yaml.Unmarshal(marshalled, &recipe)
	return
}
