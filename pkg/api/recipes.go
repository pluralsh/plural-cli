package api

import (
	"fmt"

	"github.com/pluralsh/gqlclient"
	"github.com/pluralsh/gqlclient/pkg/utils"

	"sigs.k8s.io/yaml"
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

func (client *client) CreateRecipe(repoName string, attrs gqlclient.RecipeAttributes) (string, error) {
	if len(attrs.Tests) == 0 {
		attrs.Tests = make([]*gqlclient.RecipeTestAttributes, 0)
	}
	resp, err := client.pluralClient.CreateRecipe(client.ctx, repoName, attrs)
	if err != nil {
		return "", err
	}

	return resp.CreateRecipe.ID, err
}

func (client *client) GetRecipe(repo, name string) (*Recipe, error) {
	resp, err := client.pluralClient.GetRecipe(client.ctx, &repo, &name)
	if err != nil {
		return nil, err
	}

	r := &Recipe{
		Id:                 resp.Recipe.ID,
		Name:               resp.Recipe.Name,
		Provider:           string(*resp.Recipe.Provider),
		Description:        utils.ConvertStringPointer(resp.Recipe.Description),
		Tests:              []*RecipeTest{},
		RecipeSections:     []*RecipeSection{},
		RecipeDependencies: []*Recipe{},
	}
	if resp.Recipe.OidcSettings != nil {
		r.OidcSettings = &OIDCSettings{
			DomainKey:  utils.ConvertStringPointer(resp.Recipe.OidcSettings.DomainKey),
			UriFormat:  utils.ConvertStringPointer(resp.Recipe.OidcSettings.URIFormat),
			UriFormats: utils.ConvertStringArrayPointer(resp.Recipe.OidcSettings.URIFormats),
			AuthMethod: string(resp.Recipe.OidcSettings.AuthMethod),
		}
		if resp.Recipe.OidcSettings.Subdomain != nil {
			r.OidcSettings.Subdomain = *resp.Recipe.OidcSettings.Subdomain
		}
	}
	if resp.Recipe.Repository != nil {
		r.Repository = &Repository{
			Id:   resp.Recipe.Repository.ID,
			Name: resp.Recipe.Repository.Name,
		}
	}
	if resp.Recipe.Restricted != nil {
		r.Restricted = *resp.Recipe.Restricted
	}

	for _, dep := range resp.Recipe.RecipeDependencies {
		r.RecipeDependencies = append(r.RecipeDependencies, convertRecipe(dep))
	}

	for _, section := range resp.Recipe.RecipeSections {
		rs := &RecipeSection{
			Id: fmt.Sprint(section.Index),
			Repository: &Repository{
				Id:          section.Repository.ID,
				Name:        section.Repository.Name,
				Description: utils.ConvertStringPointer(section.Repository.Description),
				Icon:        utils.ConvertStringPointer(section.Repository.Icon),
				DarkIcon:    utils.ConvertStringPointer(section.Repository.DarkIcon),
				Notes:       utils.ConvertStringPointer(section.Repository.Notes),
			},
			RecipeItems:   []*RecipeItem{},
			Configuration: []*ConfigurationItem{},
		}
		for _, conf := range section.Configuration {
			rs.Configuration = append(rs.Configuration, convertConfigurationItem(conf))
		}
		for _, recipeItem := range section.RecipeItems {
			rs.RecipeItems = append(rs.RecipeItems, convertRecipeItem(recipeItem))
		}

		r.RecipeSections = append(r.RecipeSections, rs)

	}

	for _, test := range resp.Recipe.Tests {
		t := &RecipeTest{
			Name:    test.Name,
			Type:    string(test.Type),
			Message: utils.ConvertStringPointer(test.Message),
			Args:    []*TestArgument{},
		}
		for _, arg := range test.Args {
			t.Args = append(t.Args, &TestArgument{
				Name: arg.Name,
				Repo: arg.Repo,
				Key:  arg.Key,
			})
		}

		r.Tests = append(r.Tests, t)
	}

	return r, nil
}

func convertRecipeItem(item *gqlclient.RecipeItemFragment) *RecipeItem {
	ri := &RecipeItem{
		Id:        utils.ConvertStringPointer(item.ID),
		Terraform: convertTerraform(item.Terraform),
	}
	for _, conf := range item.Configuration {
		ri.Configuration = append(ri.Configuration, convertConfigurationItem(conf))
	}
	if item.Chart != nil {
		ri.Chart = &Chart{
			Id:            utils.ConvertStringPointer(item.Chart.ID),
			Name:          item.Chart.Name,
			Description:   utils.ConvertStringPointer(item.Chart.Description),
			LatestVersion: utils.ConvertStringPointer(item.Chart.LatestVersion),
		}
	}

	return ri
}

func convertConfigurationItem(conf *gqlclient.RecipeConfigurationFragment) *ConfigurationItem {
	confItem := &ConfigurationItem{
		Name:          utils.ConvertStringPointer(conf.Name),
		Default:       utils.ConvertStringPointer(conf.Default),
		Documentation: utils.ConvertStringPointer(conf.Documentation),
		Placeholder:   utils.ConvertStringPointer(conf.Placeholder),
		FunctionName:  utils.ConvertStringPointer(conf.FunctionName),
	}
	if conf.Optional != nil {
		confItem.Optional = *conf.Optional
	}
	if conf.Type != nil {
		confItem.Type = string(*conf.Type)
	}
	if conf.Condition != nil {
		confItem.Condition = &Condition{
			Field:     conf.Condition.Field,
			Value:     utils.ConvertStringPointer(conf.Condition.Value),
			Operation: string(conf.Condition.Operation),
		}
	}
	if conf.Validation != nil {
		confItem.Validation = &Validation{
			Type:    string(conf.Validation.Type),
			Regex:   utils.ConvertStringPointer(conf.Validation.Regex),
			Message: conf.Validation.Message,
		}
	}

	return confItem
}

func convertRecipe(rcp *gqlclient.RecipeFragment) *Recipe {
	r := &Recipe{
		Id:                 rcp.ID,
		Name:               rcp.Name,
		Description:        utils.ConvertStringPointer(rcp.Description),
		Tests:              []*RecipeTest{},
		RecipeSections:     []*RecipeSection{},
		RecipeDependencies: []*Recipe{},
	}
	if rcp.Repository != nil {
		r.Repository = &Repository{
			Id:   rcp.Repository.ID,
			Name: rcp.Repository.Name,
		}
	}
	if rcp.OidcSettings != nil {
		r.OidcSettings = &OIDCSettings{
			DomainKey:  utils.ConvertStringPointer(rcp.OidcSettings.DomainKey),
			UriFormat:  utils.ConvertStringPointer(rcp.OidcSettings.URIFormat),
			UriFormats: utils.ConvertStringArrayPointer(rcp.OidcSettings.URIFormats),
			AuthMethod: string(rcp.OidcSettings.AuthMethod),
		}
	}
	if rcp.Restricted != nil {
		r.Restricted = *rcp.Restricted
	}
	if rcp.Provider != nil {
		provider := *rcp.Provider
		r.Provider = string(provider)
	}

	for _, test := range rcp.Tests {
		t := &RecipeTest{
			Name:    test.Name,
			Type:    string(test.Type),
			Message: utils.ConvertStringPointer(test.Message),
			Args:    []*TestArgument{},
		}
		for _, arg := range test.Args {
			t.Args = append(t.Args, &TestArgument{
				Name: arg.Name,
				Repo: arg.Repo,
				Key:  arg.Key,
			})
		}
		r.Tests = append(r.Tests, t)
	}

	for _, section := range rcp.RecipeSections {
		rs := &RecipeSection{
			Id: fmt.Sprint(section.Index),
			Repository: &Repository{
				Id:          section.Repository.ID,
				Name:        section.Repository.Name,
				Description: utils.ConvertStringPointer(section.Repository.Description),
				Icon:        utils.ConvertStringPointer(section.Repository.Icon),
				DarkIcon:    utils.ConvertStringPointer(section.Repository.DarkIcon),
				Notes:       utils.ConvertStringPointer(section.Repository.Notes),
			},
			RecipeItems:   []*RecipeItem{},
			Configuration: []*ConfigurationItem{},
		}
		for _, conf := range section.Configuration {
			rs.Configuration = append(rs.Configuration, convertConfigurationItem(conf))
		}
		for _, recipeItem := range section.RecipeItems {
			rs.RecipeItems = append(rs.RecipeItems, convertRecipeItem(recipeItem))
		}

		r.RecipeSections = append(r.RecipeSections, rs)

	}

	return r
}

func convertStack(st *gqlclient.StackFragment) *Stack {
	return &Stack{
		Name:        st.Name,
		Description: utils.ConvertStringPointer(st.Description),
		Featured:    *st.Featured,
	}
}

func (client *client) ListRecipes(repo, provider string) ([]*Recipe, error) {
	recipes := make([]*Recipe, 0)
	
	if provider != "" {
		p := gqlclient.Provider(NormalizeProvider(provider))
		resp, err := client.pluralClient.ListRecipes(client.ctx, &repo, &p)
		if err != nil {
			return nil, err
		}
		for _, edge := range resp.Recipes.Edges {
			recipes = append(recipes, convertRecipe(edge.Node))
		}
	} else {
		resp, err := client.pluralClient.ListAllRecipes(client.ctx, &repo)
		if err != nil {
			return nil, err
		}
		for _, edge := range resp.Recipes.Edges {
			recipes = append(recipes, convertRecipe(edge.Node))
		}
	}

	return recipes, nil
}

func (client *client) InstallRecipe(id string) error {
	_, err := client.pluralClient.InstallRecipe(client.ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (client *client) GetStack(name, provider string) (*Stack, error) {
	p := gqlclient.Provider(NormalizeProvider(provider))
	resp, err := client.pluralClient.GetStack(client.ctx, name, p)
	if err != nil {
		return nil, err
	}

	s := convertStack(resp.Stack)
	s.Bundles = make([]*Recipe, 0)
	for _, r := range resp.Stack.Bundles {
		s.Bundles = append(s.Bundles, convertRecipe(r))
	}

	return s, nil
}

func (client *client) ListStacks(featured bool) ([]*Stack, error) {
	resp, err := client.pluralClient.ListStacks(client.ctx, &featured, nil)
	if err != nil {
		return nil, err
	}

	stacks := make([]*Stack, 0)
	for _, edge := range resp.Stacks.Edges {
		stacks = append(stacks, convertStack(edge.Node))
	}

	return stacks, nil
}

func (client *client) CreateStack(attributes gqlclient.StackAttributes) (string, error) {
	resp, err := client.pluralClient.CreateStack(client.ctx, attributes)
	if err != nil {
		return "", err
	}

	return resp.CreateStack.ID, err
}

func ConstructStack(marshalled []byte) (stack gqlclient.StackAttributes, err error) {
	err = yaml.Unmarshal(marshalled, &stack)
	return
}

func ConstructRecipe(marshalled []byte) (recipe gqlclient.RecipeAttributes, err error) {
	err = yaml.Unmarshal(marshalled, &recipe)
	return
}
