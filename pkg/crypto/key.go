package crypto

import (
	"fmt"
	"encoding/base64"
	"io"
	"path/filepath"
	"os"
	"io/ioutil"
	"github.com/pluralsh/plural/pkg/utils"
	"gopkg.in/yaml.v2"
)

type KeyProvider struct {
	key string
}

func (prov *KeyProvider) SymmetricKey() ([]byte, error) {
	return base64.StdEncoding.DecodeString(prov.key)
}

func (prov *KeyProvider) ID() string {
	return "SHA256:" + sha256.Sum256(prov.key)
}

func (prov *KeyProvider) Marshall() ([]byte, error) {
	conf := Config{
		Version: "crypto.plural.sh/v1",
		Type: RAW,
		Id: prov.ID(),
		Context: map[string]interface{}{}
	}

	return yaml.Marshal(k)
}

func buildKeyProvider(conf *Config) (prov *KeyProvider, err error) {
	key, err := Materialize()
	if err != nil {
		return
	}

	prov = &KeyProvider{key: key.Key}
	if prov.ID() != conf.Id {
		err = fmt.Errorf("the key fingerprints failed to match")
	}

	return
}

// Configuration for storage and creation of raw aes symmetric keys in plural config

type AESKey struct {
	Key string
}

func Materialize() (*AESKey, error) {
	p := getKeyPath()
	if utils.Exists(p) {
		contents, err := ioutil.ReadFile(p)
		if err != nil {
			return nil, err
		}

		conf := AESKey{}
		err = yaml.Unmarshal(contents, &conf)
		if err != nil {
			return nil, err
		}

		return &conf, nil
	}

	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}

	aeskey := &AESKey{base64.StdEncoding.EncodeToString(key)}
	return aeskey, aeskey.Flush()
}

func getKeyPath() string {
	folder, _ := os.UserHomeDir()
	return filepath.Join(folder, ".plural", "key")
}

func Import(buf []byte) (*AESKey, error) {
	key := AESKey{}
	err := yaml.Unmarshal(buf, &key)
	return &key, err
}

func (k *AESKey) Marshal() ([]byte, error) {
	return yaml.Marshal(k)
}

func (k *AESKey) Flush() error {
	io, err := k.Marshal()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(getKeyPath(), io, 0644)
}