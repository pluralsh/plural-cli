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

type Installation struct {
	Repository *Repository
	User       *User
	License    string
	Context    map[string]interface{}
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

type ConfigurationItem struct {
	Name    string
	Default string
	Type    string
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
		publisher { name }
	}
`)

var InstallationFragment = fmt.Sprintf(`
	fragment InstallationFragment on Installation {
		id
		context
		license
		repository { ...RepositoryFragment }
	}
	%s
`, RepositoryFragment)

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