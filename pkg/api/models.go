package api

import (
	"fmt"
)

type PageInfo struct {
	HasNextPage bool
	EndCursor   string
}

type Publisher struct {
	Id   string
	Name string
}

type Repository struct {
	Id          string
	Name        string
	Description string
	Icon        string
	DarkIcon    string
	Publisher   *Publisher
}

type Chart struct {
	Id            string
	Name          string
	Description   string
	LatestVersion string
	Dependencies  *Dependencies
}

type ChartInstallation struct {
	Id           string
	Chart        *Chart
	Version      *Version
	Installation *Installation
}

type Tag struct {
	Tag string
}

type Version struct {
	Id             string
	Version        string
	Readme         string
	Package        string
	ValuesTemplate string
	Crds           []Crd
	Dependencies   *Dependencies
}

type Terraform struct {
	Id             string
	Name           string
	Description    string
	ValuesTemplate string
	Dependencies   *Dependencies
	Package        string
}

type Dependencies struct {
	Dependencies    []*Dependency
	Providers       []string
	Wirings         *Wirings
	Secrets         []string
	Application     bool
	ProviderWirings map[string]interface{}
	Outputs         map[string]string
}

type Dependency struct {
	Type string
	Repo string
	Name string
}

type Wirings struct {
	Terraform map[string]string
	Helm      map[string]string
}

type TerraformInstallation struct {
	Id           string
	Installation *Installation
	Terraform    *Terraform
	Version      *Version
}

type OAuthConfiguration struct {
	Issuer                string
	AuthorizationEndpoint string
	TokenEndpoint         string
	JwksUri               string
	UserinfoEndpoint      string
}

type OIDCProvider struct {
	ID            string
	ClientId      string
	ClientSecret  string
	RedirectUris  []string
	Configuration *OAuthConfiguration
}

type Installation struct {
	Repository    *Repository
	User          *User
	OIDCProvider *OIDCProvider `json:"oidcProvider"`
	License       string
	Context       map[string]interface{}
	AcmeKeyId     string
	AcmeSecret    string
}

type InstallationEdge struct {
	Node *Installation
}

type ChartEdge struct {
	Node *Chart
}

type TerraformEdge struct {
	Node *Terraform
}

type VersionEdge struct {
	Node *Version
}

type ChartInstallationEdge struct {
	Node *ChartInstallation
}

type TerraformInstallationEdge struct {
	Node *TerraformInstallation
}

type Token struct {
	Token string
}

type Webhook struct {
	Id     string
	Url    string
	Secret string
}

type Recipe struct {
	Id             string
	Name           string
	Provider       string
	Description    string
	RecipeSections []*RecipeSection
}

type RecipeSection struct {
	Id          string
	Repository  *Repository
	RecipeItems []*RecipeItem
}

type RecipeItem struct {
	Id            string
	Terraform     *Terraform
	Chart         *Chart
	Configuration []*ConfigurationItem
}

type Condition struct {
	Field     string
	Operation string
	Value     string
}

type ConfigurationItem struct {
	Name    string
	Default string
	Documentation string
	Type    string
	Placeholder string
	Condition   *Condition
}

type Artifact struct {
	Id       string
	Name     string
	Readme   string
	Blob     string
	Sha      string
	Platform string
	Arch     string
	Filesize int
}

type Crd struct {
	Id   string
	Name string
	Blob string
}

type ChartName struct {
	Repo  string
	Chart string
}

type Upgrade struct {
	Id string
}

type User struct {
	Id    string
	Email string
	Name  string
}

type PublicKey struct {
	Id string
	Content string
	User *User
}

type PublicKeyEdge struct {
	Node *PublicKey
}

var RepositoryFragment = fmt.Sprintf(`
	fragment RepositoryFragment on Repository {
		id
		name
		description
		icon
		darkIcon
		publisher { name }
	}
`)

const OIDCFragment = `
	fragment OIDCProvider on OidcProvider {
		id
		clientId
		clientSecret
		redirectUris
		configuration {
			issuer
      authorizationEndpoint
      tokenEndpoint
      jwksUri
      userinfoEndpoint
		}
	}
`

var InstallationFragment = fmt.Sprintf(`
	fragment InstallationFragment on Installation {
		id
		context
		license
		acmeKeyId
		acmeSecret
		repository { ...RepositoryFragment }
		oidcProvider { ...OIDCProvider }
	}
	%s %s
`, RepositoryFragment, OIDCFragment)

const ChartFragment = `
	fragment ChartFragment on Chart {
		id
		name
		description
		latestVersion
	}
`

const CrdFragment = `
	fragment CrdFragment on Crd {
		id
		name
		blob
	}
`

const DependenciesFragment = `
	fragment DependenciesFragment on Dependencies {
		dependencies {
			type
			name
			repo
		}
		application
		providers
		secrets
		wirings { terraform helm }
		providerWirings
		outputs
	}
`

var VersionFragment = fmt.Sprintf(`
	fragment VersionFragment on Version {
		id
		readme
		version
		valuesTemplate
		package
		crds { ...CrdFragment }
		dependencies { ...DependenciesFragment }
	}
	%s
`, CrdFragment)

var ChartInstallationFragment = fmt.Sprintf(`
	fragment ChartInstallationFragment on ChartInstallation {
		id
		chart {
			...ChartFragment
			dependencies { ...DependenciesFragment }
		}
		version { ...VersionFragment }
	}
	%s
	%s
	%s
`, ChartFragment, DependenciesFragment, VersionFragment)

var TerraformFragment = fmt.Sprintf(`
	fragment TerraformFragment on Terraform {
		id
		name
		package
		description
		dependencies { ...DependenciesFragment }
		valuesTemplate
	}
	%s
`, DependenciesFragment)

var TerraformInstallationFragment = fmt.Sprintf(`
	fragment TerraformInstallationFragment on TerraformInstallation {
		id
		terraform { ...TerraformFragment }
		version { ...VersionFragment }
	}
	%s
	%s
`, TerraformFragment, VersionFragment)

const TokenFragment = `
	fragment TokenFragment on PersistedToken {
		token
	}
`

const WebhookFragment = `
	fragment WebhookFragment on Webhook {
		id
		url
		secret
	}
`

const ArtifactFragment = `
	fragment ArtifactFragment on Artifact {
		id
		name
		readme
		platform
		arch
		blob
		sha
		filesize
	}
`

const UserFragment = `
	fragment UserFragment on User {
		id
		name
		email
	}
`

var PublicKeyFragment = fmt.Sprintf(`
	fragment PublicKeyFragment on PublicKey {
		id
		content
		user { ...UserFragment }
	}
	%s
`, UserFragment)

const RecipeFragment = `
	fragment RecipeFragment on Recipe {
		id
    name
    description
    provider
	}
`

var RecipeItemFragment = fmt.Sprintf(`
	fragment RecipeItemFragment on RecipeItem {
		id
		chart { ...ChartFragment }
		terraform { ...TerraformFragment }
		configuration {
			name
			type
			default
			documentation
			placeholder
			condition { field operation value }
		}
	}
	%s
	%s
`, ChartFragment, TerraformFragment)

var RecipeSectionFragment = fmt.Sprintf(`
fragment RecipeSectionFragment on RecipeSection {
	index
	repository { ...RepositoryFragment }
	recipeItems { ...RecipeItemFragment }
}
%s
%s
`, RepositoryFragment, RecipeItemFragment)