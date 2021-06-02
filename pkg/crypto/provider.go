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
	AWS IdentityType = "aws"
	GCP IdentityType = "gcp"
	AZ  IdentityType = "azure"
)

func (prov Provider) Encrypt(text []byte) ([]byte, error) {
	key, err := prov.SymmetricKey()
	if err != nil {
		return nil, err
	}

	return encrypt(key, text)
}

func (prov Provider) Decrypt(text []byte) ([]byte, error) {
	key, err := prov.SymmetricKey()
	if err != nil {
		return nil, err
	}

	return decrypt(key, text)
}

func (prov Provider) Flush() error {
	io, err := prov.Marshall()
	if err != nil {
		return err
	}

	p, err := getConfigPath()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(p, io, 0644)
}