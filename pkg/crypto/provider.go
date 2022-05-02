package crypto

import (
	"io/ioutil"
)

type IdentityType string

type Provider interface {
	ID() string
	SymmetricKey() ([]byte, error)
	Marshall() ([]byte, error)
}

const (
	KEY IdentityType = "key"
	AGE IdentityType = "age"
)

func Encrypt(prov Provider, text []byte) ([]byte, error) {
	key, err := prov.SymmetricKey()
	if err != nil {
		return nil, err
	}

	return encrypt(key, text)
}

func Decrypt(prov Provider, text []byte) ([]byte, error) {
	key, err := prov.SymmetricKey()
	if err != nil {
		return nil, err
	}

	return decrypt(key, text)
}

func Flush(prov Provider) error {
	io, err := prov.Marshall()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(configPath(), io, 0644)
}
