package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"golang.org/x/text/language"

	"golang.org/x/text/cases"
	"gopkg.in/oleiade/reflections.v1"
	"gopkg.in/yaml.v2"
)

const (
	pluralDir  = ".plural"
	ConfigName = "config.yml"
)

type Metadata struct {
	Name string `yaml:"name"`
}

type Config struct {
	Email           string `json:"email"`
	Token           string `yaml:"token" json:"token"`
	NamespacePrefix string `yaml:"namespacePrefix"`
	Endpoint        string `yaml:"endpoint"`
	LockProfile     string `yaml:"lockProfile"`
	ReportErrors    bool   `yaml:"reportErrors"`
	metadata        *Metadata
}

type VersionedConfig struct {
	ApiVersion string    `yaml:"apiVersion"`
	Kind       string    `yaml:"kind"`
	Metadata   *Metadata `yaml:"metadata"`
	Spec       *Config   `yaml:"spec"`
}

func configFile() string {
	folder, _ := os.UserHomeDir()
	return path.Join(folder, pluralDir, ConfigName)
}

func Exists() bool {
	_, err := os.Stat(configFile())
	return !os.IsNotExist(err)
}

func Read() Config {
	return Import(configFile())
}

func Profile(name string) error {
	folder, _ := os.UserHomeDir()
	conf := Import(path.Join(folder, pluralDir, name+".yml"))
	return conf.Flush()
}

func Profiles() ([]*VersionedConfig, error) {
	folder, _ := os.UserHomeDir()
	confDir := path.Join(folder, pluralDir)
	files, err := ioutil.ReadDir(confDir)
	confs := []*VersionedConfig{}
	if err != nil {
		return confs, err
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ConfigName) && strings.HasSuffix(f.Name(), ".yml") {
			contents, err := ioutil.ReadFile(path.Join(confDir, f.Name()))
			if err != nil {
				return confs, err
			}

			versioned := &VersionedConfig{}
			if err = yaml.Unmarshal(contents, versioned); err != nil {
				return nil, err
			}
			confs = append(confs, versioned)
		}
	}

	return confs, nil
}

func Import(file string) (conf Config) {
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	versioned := &VersionedConfig{Spec: &conf}
	if err = yaml.Unmarshal(contents, versioned); err != nil {
		return Config{}
	}
	conf.metadata = versioned.Metadata
	return
}

func FromToken(token string) error {
	conf := &Config{Token: token}
	return conf.Flush()
}

func Amend(key string, value string) error {
	key = cases.Title(language.Und, cases.NoLower).String(key)
	conf := Read()
	if err := reflections.SetField(&conf, key, value); err != nil {
		return err
	}
	return conf.Flush()
}

func (conf *Config) Marshal() ([]byte, error) {
	versioned := &VersionedConfig{
		ApiVersion: "platform.plural.sh/v1alpha1",
		Kind:       "Config",
		Spec:       conf,
		Metadata:   conf.metadata,
	}
	return yaml.Marshal(&versioned)
}

func (c *Config) Namespace(ns string) string {
	if len(c.NamespacePrefix) > 0 {
		return fmt.Sprintf("%s%s", c.NamespacePrefix, ns)
	}

	return ns
}

func (c *Config) Url() string {
	return c.BaseUrl() + "/gql"
}

func PluralUrl(endpoint string) string {
	host := "https://app.plural.sh"
	if endpoint != "" {
		host = fmt.Sprintf("https://%s", endpoint)
	}
	return host
}

func (c *Config) BaseUrl() string {
	return PluralUrl(c.Endpoint)
}

func (c *Config) SaveProfile(name string) error {
	c.metadata = &Metadata{Name: name}
	return c.Save(fmt.Sprintf("%s.yml", name))
}

func (c *Config) Save(filename string) error {
	io, err := c.Marshal()
	if err != nil {
		return err
	}

	folder, _ := os.UserHomeDir()
	if err := os.MkdirAll(path.Join(folder, pluralDir), os.ModePerm); err != nil {
		return err
	}

	return ioutil.WriteFile(path.Join(folder, pluralDir, filename), io, 0644)
}

func (c *Config) Flush() error {
	return c.Save(ConfigName)
}
