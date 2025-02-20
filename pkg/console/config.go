package console

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"
)

const (
	pluralDir  = ".plural"
	ConfigName = "console.yml"
)

type VersionedConfig struct {
	ApiVersion string  `json:"apiVersion"`
	Kind       string  `json:"kind"`
	Spec       *Config `json:"spec"`
}

type Config struct {
	Url   string `json:"url"`
	Token string `json:"token"`
}

func configFile() string {
	folder, _ := os.UserHomeDir()
	return filepath.Join(folder, pluralDir, ConfigName)
}

func ReadConfig() (conf Config) {
	contents, err := os.ReadFile(configFile())
	if err != nil {
		return
	}

	versioned := &VersionedConfig{Spec: &conf}
	if err = yaml.Unmarshal(contents, versioned); err != nil {
		return Config{}
	}
	return
}

func (conf *Config) Validate() error {
	url, err := url.Parse(conf.Url)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}

	if url.Scheme != "https" {
		return fmt.Errorf("url scheme must be https")
	}

	return nil
}

func (conf *Config) Save() error {
	if err := conf.Validate(); err != nil {
		return err
	}

	versioned := &VersionedConfig{
		ApiVersion: "platform.plural.sh/v1alpha1",
		Kind:       "Console",
		Spec:       conf,
	}
	io, err := yaml.Marshal(versioned)
	if err != nil {
		return err
	}

	f := configFile()
	if err := os.MkdirAll(filepath.Dir(f), os.ModePerm); err != nil {
		return err
	}

	return os.WriteFile(f, io, 0644)
}
