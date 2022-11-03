package crypto

import (
	"os"
	"path/filepath"

	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Version string
	Type    IdentityType
	Id      string
	Context map[string]interface{}
}

func configPath() string {
	root, _ := utils.ProjectRoot()
	return pathing.SanitizeFilepath(filepath.Join(root, "crypto.yml"))
}

func ReadConfig() (conf *Config, err error) {
	conf = &Config{}
	contents, err := os.ReadFile(configPath())
	if err != nil {
		return
	}

	err = yaml.Unmarshal(contents, &conf)
	return
}

func Build() (Provider, error) {
	if utils.Exists(configPath()) {
		conf, err := ReadConfig()
		if err != nil {
			return fallbackProvider()
		}

		switch conf.Type {
		case KEY:
			return buildKeyProvider(conf)
		case AGE:
			return BuildAgeProvider()
		}
	}

	return fallbackProvider()
}

func fallbackProvider() (*KeyProvider, error) {
	key, err := Materialize()
	if err != nil {
		return nil, err
	}
	return &KeyProvider{key: key.Key}, err
}
