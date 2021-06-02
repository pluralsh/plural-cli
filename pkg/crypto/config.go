package crypto

import (
	"fmt"
	"path/filepath"
	"encoding/base64"
	"crypto/sha256"
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

func Config() (conf *Config, err error) {
	conf = &Config{}
	contents, err := ioutil.ReadFile(configPath())
	if err != nil {
		return
	}

	err = yaml.Unmarshal(contents, &conf)
	return
}

func Build() (prov Provider, err error) {
	if utils.Exists(configPath()) {
		conf, err := Config()
		switch (conf.Type) {
		case KEY:
			prov, err = buildKeyProvider(conf)
			return
		}
	}

	key, err := Materialize()
	prov = &KeyProvider{key: key.Key}
	return
}