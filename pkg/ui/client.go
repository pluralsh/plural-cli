//go:build ui || generate

package ui

import (
	"encoding/json"
	"fmt"

	"github.com/pluralsh/polly/algorithms"
	"github.com/urfave/cli"
	"golang.org/x/exp/maps"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/bundle"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/manifest"
)

type Application struct {
	Key          string `json:"key"`
	Label        string `json:"label"`
	IsDependency bool   `json:"isDependency"`
	// DependencyOf is a set of application names that this app is a dependency of.
	DependencyOf map[string]interface{} `json:"dependencyOf"`
	Data         map[string]interface{} `json:"data"`
}

func (this *Application) UnmarshalJSON(data []byte) error {
	type Alias struct {
		Key          string `json:"key"`
		Label        string `json:"label"`
		IsDependency bool   `json:"isDependency"`
		// Since Set does not exist in Go, we are passing array
		// from the frontend and converting it to a map.
		DependencyOf []string               `json:"dependencyOf"`
		Data         map[string]interface{} `json:"data"`
	}
	alias := &Alias{}
	if err := json.Unmarshal(data, alias); err != nil {
		return fmt.Errorf("error during Application.UnmarshalJSON: %v\n", err)
	}

	dependencyOf := map[string]interface{}{}
	for _, appName := range alias.DependencyOf {
		dependencyOf[appName] = struct{}{}
	}

	*this = Application{
		Key:          alias.Key,
		Label:        alias.Label,
		IsDependency: alias.IsDependency,
		DependencyOf: dependencyOf,
		Data:         alias.Data,
	}

	return nil
}

// Client struct used by the frontend to access and update backend data.
type Client struct {
	ctx    *cli.Context
	client api.Client
}

func (this *Client) Token() string {
	conf := config.Read()

	return conf.Token
}

func (this *Client) Project() *manifest.ProjectManifest {
	project, err := manifest.FetchProject()
	if err != nil {
		return nil
	}

	return project
}

func (this *Client) Context() *manifest.Context {
	context, err := manifest.FetchContext()
	if err != nil {
		return nil
	}

	return context
}

func (this *Client) Install(applications []Application, domains, buckets []string) error {
	path := manifest.ContextPath()
	context, err := manifest.ReadContext(path)
	if err != nil {
		context = manifest.NewContext()
	}

	this.addDomains(context, domains)
	this.addBuckets(context, buckets)

	installableApplications := algorithms.Filter(applications, func(app Application) bool {
		return !app.IsDependency
	})

	dependencies := algorithms.Filter(applications, func(app Application) bool {
		return app.IsDependency
	})

	for _, dep := range dependencies {
		if err = this.doInstall(dep, context); err != nil {
			return err
		}
	}

	for _, app := range installableApplications {
		if err = this.doInstall(app, context); err != nil {
			return err
		}
	}

	// Write to context.yaml only if there were no errors
	err = context.Write(path)
	return err
}

func (this *Client) doInstall(application Application, context *manifest.Context) error {
	recipeID := application.Data["id"].(string)
	oidc := application.Data["oidc"].(bool)
	configuration := application.Data["context"].(map[string]interface{})
	repoName := application.Label
	mergedConfiguration, exists := context.Configuration[repoName]
	if !exists {
		mergedConfiguration = map[string]interface{}{}
	}

	recipe, err := this.client.GetRecipeByID(recipeID)
	if err != nil {
		return api.GetErrorResponse(err, "GetRecipeByID")
	}

	// Merge incoming configuration with existing one and update context
	maps.Copy(mergedConfiguration, configuration)
	context.Configuration[repoName] = mergedConfiguration

	// Non-dependency apps need some additional handling
	if !application.IsDependency {
		// Add installed app to the context
		context.AddBundle(repoName, recipe.Name)

		// Install app recipe
		//if err := this.client.InstallRecipe(recipeID); err != nil {
		//	return fmt.Errorf("error: %w", api.GetErrorResponse(err, "InstallRecipe"))
		//}
	}

	// Configure OIDC if enabled
	if oidc {
		confirm := false
		err = bundle.ConfigureOidc(repoName, this.client, recipe, configuration, &confirm)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *Client) addDomains(context *manifest.Context, domains []string) {
	for _, domain := range domains {
		if !context.HasDomain(domain) {
			context.AddDomain(domain)
		}
	}
}

func (this *Client) addBuckets(context *manifest.Context, buckets []string) {
	for _, bucket := range buckets {
		if !context.HasBucket(bucket) {
			context.AddBucket(bucket)
		}
	}
}

// NewClient creates a new proxy client struct
func NewClient(client api.Client, ctx *cli.Context) *Client {
	return &Client{
		ctx:    ctx,
		client: client,
	}
}
