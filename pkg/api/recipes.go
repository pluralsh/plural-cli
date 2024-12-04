package api

type RecipeInput struct {
	Name         string
	Description  string
	Provider     string
	Restricted   bool
	Primary      bool
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
