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

func Build() (prov Provider, err error) {
	key, err := Materialize()
	if err != nil {
		return
	}
	keyID, err := GetKeyID()
	if err != nil {
		return
	}

	prov, err = fallbackProvider(key)
	if err != nil {
		return
	}
	if utils.Exists(configPath()) {
		var conf *Config
		conf, err = ReadConfig()
		if err != nil {
			return fallbackProvider(key)
		}

		switch conf.Type {
		case KEY:
			prov, err = buildKeyProvider(conf, key)
		case AGE:
			prov, err = BuildAgeProvider()
		}
		if err != nil {
			return
		}
	}

	if keyID != "" && prov.ID() != keyID {
		err = errFingerprint
	}

	return
}

func fallbackProvider(key *AESKey) (*KeyProvider, error) {
	return &KeyProvider{key: key.Key}, nil
}
