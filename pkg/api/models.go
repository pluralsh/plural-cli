package api

import (
	"github.com/pluralsh/gqlclient"
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
	Helm           map[string]interface{}
	Package        string
	ValuesTemplate string
	TemplateType   gqlclient.TemplateType
	Crds           []Crd
	Dependencies   *Dependencies
	InsertedAt     string
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
	CliVsn          string
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
	Primary            bool
	Restricted         bool
	Tests              []*RecipeTest
	Repository         *Repository
	RecipeSections     []*RecipeSection
	OidcSettings       *OIDCSettings `yaml:"oidcSettings" json:"oidcSettings"`
	RecipeDependencies []*Recipe     `yaml:"recipeDependencies" json:"recipeDependencies"`
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
	Values        []string
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

type KeyBackup struct {
	Name         string
	Digest       string
	Repositories []string
	Value        string
	InsertedAt   string
}

type Cluster struct {
	Id          string
	Name        string
	Provider    string
	UpgradeInfo []*UpgradeInfo
	Source      string
	GitUrl      string
	Owner       *User
}

type UpgradeInfo struct {
	Count        int64
	Installation *Installation
}

type ChatMessage struct {
	Name    string
	Content string
	Role    string
}
