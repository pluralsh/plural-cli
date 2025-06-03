package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"gopkg.in/yaml.v2"
)

func init() {
	EncryptionKeyFile = ""
}

var EncryptionKeyFile string

type KeyProvider struct {
	key string
}

func (prov *KeyProvider) SymmetricKey() ([]byte, error) {
	return base64.StdEncoding.DecodeString(prov.key)
}

func (prov *KeyProvider) ID() string {
	sha := sha256.Sum256([]byte(prov.key))
	return "SHA256:" + base32.StdEncoding.EncodeToString(sha[:])
}

func (prov *KeyProvider) Marshall() ([]byte, error) {
	conf := Config{
		Version: "crypto.plural.sh/v1",
		Type:    KEY,
		Id:      prov.ID(),
	}

	return yaml.Marshal(conf)
}

func buildKeyProvider(conf *Config, key *AESKey) (prov *KeyProvider, err error) {
	if conf.Context != nil && conf.Context.Key != nil {
		if file, err := homedir.Expand(conf.Context.Key.File); err == nil {
			if k, err := Read(file); err == nil {
				key = k
			}
		}
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

func Read(path string) (*AESKey, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return DeserializeKey(contents)
}

func Materialize() (*AESKey, error) {
	p := getKeyPath()
	// if key file already exists, always try to use it
	if utils.Exists(p) {
		return Read(p)
	}

	key, err := RandStr(32)
	if err != nil {
		return nil, err
	}

	aeskey := &AESKey{key}
	return aeskey, aeskey.Flush()
}

func RandStr(length int) (string, error) {
	str := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, str); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(str), nil
}

func getKeyPath() string {
	if EncryptionKeyFile != "" {
		return EncryptionKeyFile
	}
	folder, _ := os.UserHomeDir()
	return pathing.SanitizeFilepath(filepath.Join(folder, ".plural", "key"))
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
	if err := backupKey(key); err != nil {
		return err
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

	return os.WriteFile(getKeyPath(), io, 0644)
}

func (k *AESKey) ID() string {
	sha := sha256.Sum256([]byte(k.Key))
	return "SHA256:" + base32.StdEncoding.EncodeToString(sha[:])
}

type KeyValidator struct {
	KeyID string
}

func (k *KeyValidator) Marshal() ([]byte, error) {
	return yaml.Marshal(k)
}

func (k *KeyValidator) Flush() error {
	io, err := k.Marshal()
	if err != nil {
		return err
	}

	return os.WriteFile(getKeyValidatorPath(), io, 0644)
}

func GetKeyID() (string, error) {
	path := getKeyValidatorPath()
	if !utils.Exists(path) {
		return "", nil
	}
	contents, err := utils.ReadFile(path)
	if err != nil {
		return "", err
	}
	var k KeyValidator
	err = yaml.Unmarshal([]byte(contents), &k)
	if err != nil {
		return "", err
	}
	return k.KeyID, nil
}

func CreateKeyFingerprintFile() error {
	aesKey, err := Materialize()
	if err != nil {
		return err
	}
	kv := KeyValidator{KeyID: aesKey.ID()}
	if err := kv.Flush(); err != nil {
		return err
	}
	return nil
}

func getKeyValidatorPath() string {
	root, _ := utils.ProjectRoot()
	return pathing.SanitizeFilepath(filepath.Join(root, ".keyid"))
}
