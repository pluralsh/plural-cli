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
	Notes       string
	Publisher   *Publisher
	Recipes     []*Recipe
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
	Wait            bool
	ProviderWirings map[string]interface{}
	Outputs         map[string]interface{}
	ProviderVsn     string
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
	Id            string
	ClientId      string
	ClientSecret  string
	RedirectUris  []string
	Bindings      []*ProviderBinding
	Configuration *OAuthConfiguration
}

type Installation struct {
	Id           string
	Repository   *Repository
	User         *User
	OIDCProvider *OIDCProvider `json:"oidcProvider"`
	LicenseKey   string
	Context      map[string]interface{}
	AcmeKeyId    string
	AcmeSecret   string
}

type CloudShell struct {
	Id     string
	AesKey string `json:"aesKey"`
	GitUrl string `json:"gitUrl"`
}

type RepositoryEdge struct {
	Node *Repository
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
	Id                 string
	Name               string
	Provider           string
	Description        string
	Restricted         bool
	Tests              []*RecipeTest
	Repository         *Repository
	RecipeSections     []*RecipeSection
	OidcSettings       *OIDCSettings `yaml:"oidcSettings" json:"oidcSettings"`
	RecipeDependencies []*Recipe     `yaml:"recipeDependencies" json:"recipeDependencies"`
}

type Stack struct {
	Id          string
	Name        string
	Provider    string
	Featured    bool
	Description string
	Bundles     []*Recipe
}

type RecipeTest struct {
	Name    string
	Type    string
	Message string
	Args    []*TestArgument
}

type TestArgument struct {
	Name string
	Repo string
	Key  string
}

type OIDCSettings struct {
	DomainKey  string   `yaml:"domainKey"`
	UriFormat  string   `yaml:"uriFormat"`
	UriFormats []string `yaml:"uriFormats"`
	AuthMethod string   `yaml:"authMethod"`
	Subdomain  bool     `yaml:"subdomain"`
}

type RecipeSection struct {
	Id            string
	Repository    *Repository
	RecipeItems   []*RecipeItem
	Configuration []*ConfigurationItem
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

type Validation struct {
	Type    string
	Regex   string
	Message string
}

type ConfigurationItem struct {
	Name          string
	Default       string
	Documentation string
	Type          string
	Placeholder   string
	FunctionName  string `json:"functionName" yaml:"functionName"`
	Optional      bool
	Condition     *Condition
	Validation    *Validation
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

type Group struct {
	Id   string
	Name string
}

type ProviderBinding struct {
	User  *User
	Group *Group
}

type PublicKey struct {
	Id      string
	Content string
	User    *User
}

type PublicKeyEdge struct {
	Node *PublicKey
}

type EabCredential struct {
	Id       string
	KeyId    string
	HmacKey  string
	Cluster  string
	Provider string
}

type DnsDomain struct {
	Id   string
	Name string
}

type ApplyLock struct {
	Id   string
	Lock string
}

type ScaffoldFile struct {
	Path    string
	Content string
}

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
		wait
		application
		providers
		secrets
		wirings { terraform helm }
		providerWirings
		outputs
		providerVsn
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
