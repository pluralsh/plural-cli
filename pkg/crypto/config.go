package crypto

import (
	"io/ioutil"
	"path/filepath"
	"gopkg.in/yaml.v2"
	"github.com/pluralsh/plural/pkg/utils"
)

type Config struct {
	Version string
	Type IdentityType
	Id string
	Context map[string]interface{}
}

func configPath() string {
	root, _ := utils.ProjectRoot()
	return filepath.Join(root, "crypto.yml")
}

func ReadConfig() (conf *Config, err error) {
	conf = &Config{}
	contents, err := ioutil.ReadFile(configPath())
	if err != nil {
		return
	}

	err = yaml.Unmarshal(contents, &conf)
	return
}

func Build() (Provider, error) {
	fallback, err := fallbackProvider()
	if utils.Exists(configPath()) {
		conf, err := ReadConfig()

		if err != nil {
			return fallback, err
		}

		switch (conf.Type) {
		case KEY:
			return buildKeyProvider(conf)
		case AGE:
			return BuildAgeProvider()
		}
	}

	return fallback, err
}

func fallbackProvider() (*KeyProvider, error) {
	key, err := Materialize()
	return &KeyProvider{key: key.Key}, err
}