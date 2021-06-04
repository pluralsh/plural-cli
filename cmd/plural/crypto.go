package main

import (
	"bytes"
	"github.com/pluralsh/plural/pkg/crypto"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/urfave/cli"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

var prefix = []byte("CHARTMART-ENCRYPTED")

const gitattributes = `/**/helm/**/values.yaml filter=plural-crypt diff=plural-crypt
/**/manifest.yaml filter=plural-crypt diff=plural-crypt
/diffs/**/* filter=plural-crypt diff=plural-crypt
.gitattributes !filter !diff
`

const gitignore = `/**/.terraform
/**/.terraform*
/**/terraform.tfstate*
/bin
`

func cryptoCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "encrypt",
			Usage:  "encrypts stdin and writes to stdout",
			Action: handleEncrypt,
		},
		{
			Name:   "decrypt",
			Usage:  "decrypts stdin and writes to stdout",
			Action: handleDecrypt,
		},
		{
			Name:   "init",
			Usage:  "initializes git filters for you",
			Action: cryptoInit,
		},
		{
			Name:   "unlock",
			Usage:  "auto-decrypts all affected files in the repo",
			Action: handleUnlock,
		},
		{
			Name:   "import",
			Usage:  "imports an aes key for plural to use",
			Action: importKey,
		},
		{
			Name:   "export",
			Usage:  "dumps the current aes key to stdout",
			Action: exportKey,
		},
		{
			Name: "share",
			Usage: "allows a list of plural users to decrypt this repository",
			ArgsUsage: "",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "email",
					Usage: "a email to share with (multiple allowed)",
				},
			},
			Action: handleCryptoShare,
		},
		{
			Name: "setup-keys",
			Usage: "creates an age keypair, and uploads the public key to plural for use in plural crypto share",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name",
					Usage: "a name for the key",
				},
			},
			Action: handleSetupKeys,
		},
	}
}

func handleEncrypt(c *cli.Context) error {
	data, err := ioutil.ReadAll(os.Stdin)
	if bytes.HasPrefix(data, prefix) {
		os.Stdout.Write(data)
		return nil
	}

	if err != nil {
		return err
	}

	prov, err := crypto.Build()
	if err != nil {
		return err
	}

	result, err := crypto.Encrypt(prov, data)
	if err != nil {
		return err
	}
	os.Stdout.Write(prefix)
	os.Stdout.Write(result)
	return nil
}

func handleDecrypt(c *cli.Context) error {
	var file io.Reader
	if c.Args().Present() {
		p, _ := filepath.Abs(c.Args().First())
		f, err := os.Open(p)
		defer f.Close()
		if err != nil {
			return err
		}
		file = f
	} else {
		file = os.Stdin
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	if !bytes.HasPrefix(data, prefix) {
		os.Stdout.Write(data)
		return nil
	}

	prov, err := crypto.Build()
	if err != nil {
		return err
	}
	
	result, err := crypto.Decrypt(prov, data[len(prefix):])
	if err != nil {
		return err
	}
	
	os.Stdout.Write(result)
	return nil
}

func cryptoInit(c *cli.Context) error {
	encryptConfig := [][]string{
		{"filter.plural-crypt.smudge", "plural crypto decrypt"},
		{"filter.plural-crypt.clean", "plural crypto encrypt"},
		{"filter.plural-crypt.required", "true"},
		{"diff.plural-crypt.textconv", "plural crypto decrypt"},
	}

	utils.Highlight("Creating git encryption filters\n\n")
	for _, conf := range encryptConfig {
		if err := gitConfig(conf[0], conf[1]); err != nil {
			return err
		}
	}

	utils.WriteFileIfNotPresent(".gitattributes", gitattributes)
	utils.WriteFileIfNotPresent(".gitignore", gitignore)
	return nil
}

func handleCryptoShare(c *cli.Context) error {
	emails := c.StringSlice("email")
	if err := crypto.SetupAge(emails); err != nil {
		return err
	}

	prov, err := crypto.BuildAgeProvider()
	if err != nil {
		return err
	}

	return crypto.Flush(prov)
}

func handleSetupKeys(c *cli.Context) error {
	if err := crypto.SetupIdentity(c.String("name")); err != nil {
		return err
	}

	utils.Success("Public key uploaded successfully\n")
	return nil
}

func handleUnlock(c *cli.Context) error {
	repoRoot, err := utils.RepoRoot()
	if err != nil {
		return err
	}

	gitIndex, _ := filepath.Abs(filepath.Join(repoRoot, ".git", "index"))
	err = os.Remove(gitIndex)
	if err != nil {
		return err
	}

	return gitCommand("checkout", "HEAD", "--", repoRoot).Run()
}

func exportKey(c *cli.Context) error {
	key, err := crypto.Materialize()
	if err != nil {
		return err
	}
	io, err := key.Marshal()
	if err != nil {
		return err
	}
	os.Stdout.Write(io)
	return nil
}

func importKey(c *cli.Context) error {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	key, err := crypto.Import(data)
	if err != nil {
		return err
	}
	return key.Flush()
}
