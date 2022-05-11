package scm

import (
	"os"
	"io/ioutil"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"golang.org/x/crypto/ssh"
	"github.com/pluralsh/plural/pkg/utils"
	homedir "github.com/mitchellh/go-homedir"
	"path/filepath"
)

func generateKeys() (pub string, priv string, err error) {
	pub, priv, found := readKeys()
	if found {
		return
	}

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return
	}

	b, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return
	}

	priv = string(pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: b,
	}))

	sshPub, err := ssh.NewPublicKey(pubKey)
	if err != nil {
		return
	}
	pub = string(ssh.MarshalAuthorizedKey(sshPub))

	err = saveKeys(pub, priv)
	return
}

func readKeys() (pub string, priv string, found bool) {
	pubFile, privFile, err := keyFiles()
	if err != nil {
		return
	}

	if !utils.Exists(pubFile) || !utils.Exists(privFile) {
		return
	}

	pub, err = utils.ReadFile(pubFile)
	if err != nil {
		return
	}

	priv, err = utils.ReadFile(privFile)
	if err != nil {
		return
	}

	found = true
	return
}

func saveKeys(pub, priv string) error {
	if !utils.Confirm("Would you like to save the keys to ~/.ssh?") {
		return nil
	}

	pubFile, privFile, err := keyFiles()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(pubFile), 0700); err != nil {
		return err
	}

	if err := ioutil.WriteFile(privFile, []byte(priv), 0600); err != nil {
		return err
	}

	if err := ioutil.WriteFile(pubFile, []byte(pub), 0644); err != nil {
		return err
	}

	if sshadd, _ := utils.Which("ssh-add"); sshadd {
		return utils.Exec("ssh-add", privFile)
	}

	utils.Highlight("It looks like ssh isn't configured locally, once you have it set up, you can run `ssh-add ~/.ssh/id_plural` to add the key to your agent")
	return err
}

func keyFiles() (pub string, priv string, err error) {
	path, err := homedir.Expand("~/.ssh")
	if err != nil {
		return
	}

	pub = filepath.Join(path, "id_plural.pub")
	priv = filepath.Join(path, "id_plural")
	return
}