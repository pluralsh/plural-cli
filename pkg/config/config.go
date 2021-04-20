package config

import (
	"fmt"
	"gopkg.in/oleiade/reflections.v1"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type Config struct {
	Email string `json:"email"`
	Token string `yaml:"token" json:"token"`
	NamespacePrefix string `yaml:"namespacePrefix"`
	Endpoint string `yaml:"endpoint"`
}

type VersionedConfig struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Spec       *Config `yaml:"spec"`
}

func configFile() string {
	folder, _ := os.UserHomeDir()
	return path.Join(folder, ".plural", "config.yml")
}

func Read() Config {
	return Import(configFile())
}

func Profile(name string) error {
	folder, _ := os.UserHomeDir()
	conf := Import(path.Join(folder, ".plural", name + ".yml"))
	return conf.Flush()
}

func Import(file string) (conf Config) {
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	versioned := &VersionedConfig{Spec: &conf}
	yaml.Unmarshal(contents, versioned)
	return
}

func Amend(key string, value string) error {
	key = strings.Title(key)
	conf := Read()
	reflections.SetField(&conf, key, value)
	return conf.Flush()
}

func (conf *Config) Marshal() ([]byte, error) {
	versioned := &VersionedConfig{
		ApiVersion: "platform.plural.sh/v1alpha1",
		Kind: "Config",
		Spec: conf,
	}
	return yaml.Marshal(&versioned)
}

func (c *Config) Namespace(ns string) string {
	if (len(c.NamespacePrefix) > 0) {
		return fmt.Sprintf("%s%s", c.NamespacePrefix, ns)
	}

	return ns
}

func (c *Config) Url() string {
	return c.BaseUrl() + "/gql"
}

func (c *Config) BaseUrl() string {
	host := "https://app.plural.sh"
	if (c.Endpoint != "") {
		host = fmt.Sprintf("https://%s", c.Endpoint)
	}
	return host
}

func (c *Config) SaveProfile(name string) error {
	return c.Save(fmt.Sprintf("%s.yml", name))
}

func (c *Config) Save(filename string) error {
	io, err := c.Marshal()
	if err != nil {
		return err
	}

	folder, _ := os.UserHomeDir()
	if err := os.MkdirAll(path.Join(folder, ".plural"), os.ModePerm); err != nil {
		return err
	}

	return ioutil.WriteFile(path.Join(folder, ".plural", filename), io, 0644)
}

func (c *Config) Flush() error {
	return c.Save("config.yml")
}
