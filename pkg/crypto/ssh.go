package crypto

import (
	"crypto/ssh"
	"os"
	"io/ioutil"
	"strings"
	"path/filepath"
)

type SSHProvider struct {
	pub ssh.PublicKey
	priv interface{}
	encrypted string
}

func (id *Identity) ToSSH() (*SSHProvider, error) {
	if id.Type != SSH {
		return nil, fmt.Errorf("Not an ssh identity")
	}

	pub, priv, err := findSSHKey(id.ID)
	if err != nil {
		return nil, err
	}

	return &SSHProvider{pub: pub, priv: priv, encrypted: id.EncryptedKey}, nil
}

func findSSHKey(print string) (pub ssh.PublicKey, priv interface{}, err error) {
	folder, _ := os.UserHomeDir()
	sshKeys := make([]string)
	err = filepath.Walk(filepath.Join(folder, ".ssh"), func(path string, info os.FileInfo, err error) {
		if strings.HasSuffix(path, ".pub") {
			sshKeys = append(sshKeys, path)
		}
		return nil
	})

	if err != nil {
		return
	}

	for _, key := sshKeys {
		content, err := ioutil.ReadFile(key)
		if err != nil {
			return
		}

		pub, err = ssh.ParsePublicKey(content)
		if err != nil {
			continue
		}

		if ssh.FingerprintSHA256(pub) == print {
			privFile := strings.TrimSuffix(key, ".pub")
			content, err := ioutil.ReadFile(privFile)
			if err != nil {
				continue
			}

			priv, err = ssh.ParseRawPrivateKey(content)
			if err != nil {
				continue
			}

			return
		}
	}
	err = fmt.Errorf("No public key matching %s", print)
	return
}

