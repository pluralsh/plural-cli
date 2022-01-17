package crypto

import (
	"fmt"
	"encoding/base64"
	"crypto/rand"
	"crypto/sha256"
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
	sha := sha256.Sum256([]byte(prov.key))
	return "SHA256:" + base64.StdEncoding.EncodeToString(sha[:])
}

func (prov *KeyProvider) Marshall() ([]byte, error) {
	conf := Config{
		Version: "crypto.plural.sh/v1",
		Type: KEY,
		Id: prov.ID(),
		Context: map[string]interface{}{},
	}

	return yaml.Marshal(conf)
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

		return DeserializeKey(contents)
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

func DeserializeKey(contents []byte) (k *AESKey, err error) {
  err = yaml.Unmarshal(contents, &k)
	return
}

func Setup(key string) error {
	p := getKeyPath()
	if utils.Exists(p) {
		if err := os.Rename(p, fmt.Sprintf("%s.bak", p)); err != nil {
			return err
		}
	}

	aes := &AESKey{Key: key}
	return aes.Flush()
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