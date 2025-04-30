package manifest

import jsoniter "github.com/json-iterator/go"

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
}

type Owner struct {
	Email    string
	Endpoint string `yaml:"endpoint,omitempty"`
}

type NetworkConfig struct {
	Subdomain string `json:"subdomain"`
	PluralDns bool   `json:"pluralDns"`
}

type ProjectManifest struct {
	Cluster           string
	Bucket            string
	Project           string
	Provider          string
	Region            string
	Owner             *Owner
	Network           *NetworkConfig
	AvailabilityZones []string
	BucketPrefix      string `yaml:"bucketPrefix"`
	Context           map[string]interface{}
}

func (pm *ProjectManifest) MarshalJSON() ([]byte, error) {
	json := jsoniter.ConfigCompatibleWithStandardLibrary

	return json.Marshal(&struct {
		Cluster           string                 `json:"cluster"`
		Bucket            string                 `json:"bucket"`
		Project           string                 `json:"project"`
		Provider          string                 `json:"provider"`
		Region            string                 `json:"region"`
		Owner             *Owner                 `json:"owner"`
		Network           *NetworkConfig         `json:"network"`
		AvailabilityZones []string               `json:"availabilityZones"`
		BucketPrefix      string                 `yaml:"bucketPrefix" json:"bucketPrefix"`
		Context           map[string]interface{} `json:"context"`
	}{
		Cluster:           pm.Cluster,
		Bucket:            pm.Bucket,
		Project:           pm.Project,
		Provider:          pm.Provider,
		Region:            pm.Region,
		Owner:             pm.Owner,
		Network:           pm.Network,
		AvailabilityZones: pm.AvailabilityZones,
		BucketPrefix:      pm.BucketPrefix,
		Context:           pm.Context,
	})
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
