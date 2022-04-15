package manifest

type ChartManifest struct {
	Id        string
	Name      string
	VersionId string
	Version   string
}

type TerraformManifest struct {
	Id   string
	Name string
}

type Dependency struct {
	Repo string
}

type Metadata struct {
	Name   string
	Labels map[string]string `yaml:",omitempty"`
}

type Manifest struct {
	Id           string
	Name         string
	Cluster      string
	Project      string
	Bucket       string
	Provider     string
	Region       string
	License      string
	Wait         bool
	Charts       []*ChartManifest
	Terraform    []*TerraformManifest
	Dependencies []*Dependency
	Context      map[string]interface{}
	Links        *Links `yaml:"links,omitempty"`
}

type Owner struct {
	Email    string
	Endpoint string `yaml:"endpoint,omitempty"`
}

type NetworkConfig struct {
	Subdomain string
	PluralDns bool
}

type ProjectManifest struct {
	Cluster      string
	Bucket       string
	Project      string
	Provider     string
	Region       string
	Owner        *Owner
	Network      *NetworkConfig
	BucketPrefix string `yaml:"bucketPrefix"`
	Context      map[string]interface{}
}

type VersionedManifest struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string
	Metadata   *Metadata
	Spec       *Manifest
}

type VersionedProjectManifest struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string
	Metadata   *Metadata
	Spec       *ProjectManifest
}