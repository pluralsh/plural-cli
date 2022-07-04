package scm

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mikesmitty/edkey"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"

	"github.com/pluralsh/plural/pkg/utils"
)

type keys struct {
	pub  string
	priv string
}

func GenerateKeys(saveKeyFiles bool) (pub string, priv string, err error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return
	}

	priv = string(pem.EncodeToMemory(&pem.Block{
		Type:  "OPENSSH PRIVATE KEY",
		Bytes: edkey.MarshalED25519PrivateKey(privKey),
	}))

	sshPub, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		return
	}
	pub = string(ssh.MarshalAuthorizedKey(sshPub))

	if saveKeyFiles {
		err = saveKeys(pub, priv)
		return
	}

	return
}

func saveKeys(pub, priv string) error {
	if !utils.Confirm("Would you like to save the keys to ~/.ssh?") {
		return nil
	}

	keys, err := keyFiles()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(keys.pub), 0700); err != nil {
		return err
	}

	if err := ioutil.WriteFile(keys.priv, []byte(priv), 0600); err != nil {
		return err
	}

	if err := ioutil.WriteFile(keys.pub, []byte(pub), 0644); err != nil {
		return err
	}

	return nil
}

func keyFiles() (keys keys, err error) {
	path, err := homedir.Expand("~/.ssh")
	if err != nil {
		return
	}

	keys.pub = filepath.Join(path, "id_plural.pub")
	keys.priv = filepath.Join(path, "id_plural")
	return
}
