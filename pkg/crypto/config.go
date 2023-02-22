package crypto

import (
	"fmt"
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
	key, err := Materialize()
	if err != nil {
		return nil, err
	}
	keyID, err := GetKeyID()
	if err != nil {
		return nil, err
	}

	if err := validateKey(keyID, key); err != nil {
		return nil, err
	}

	if utils.Exists(configPath()) {
		conf, err := ReadConfig()
		if err != nil {
			return fallbackProvider(key)
		}

		switch conf.Type {
		case KEY:
			return buildKeyProvider(conf, key)
		case AGE:
			return BuildAgeProvider()
		}
	}

	return fallbackProvider(key)
}

func fallbackProvider(key *AESKey) (*KeyProvider, error) {
	return &KeyProvider{key: key.Key}, nil
}

func validateKey(keyID string, key *AESKey) error {
	if keyID != "" && key.ID() != keyID {
		return fmt.Errorf("the key fingerprint doesn't match")
	}
	return nil
}
