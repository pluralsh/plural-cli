package plural

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"

	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/crypto"
	"github.com/pluralsh/plural/pkg/scm"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
)

var prefix = []byte("CHARTMART-ENCRYPTED")

const (
	GitAttributesFile = ".gitattributes"
	GitIgnoreFile     = ".gitignore"
)

const Gitattributes = `/**/helm/**/values.yaml filter=plural-crypt diff=plural-crypt
/**/helm/**/values.yaml* filter=plural-crypt diff=plural-crypt
/**/helm/**/default-values.yaml* filter=plural-crypt diff=plural-crypt
/**/terraform/**/main.tf filter=plural-crypt diff=plural-crypt
/**/terraform/**/main.tf* filter=plural-crypt diff=plural-crypt
/**/manifest.yaml filter=plural-crypt diff=plural-crypt
/**/output.yaml filter=plural-crypt diff=plural-crypt
/diffs/**/* filter=plural-crypt diff=plural-crypt
context.yaml filter=plural-crypt diff=plural-crypt
workspace.yaml filter=plural-crypt diff=plural-crypt
context.yaml* filter=plural-crypt diff=plural-crypt
workspace.yaml* filter=plural-crypt diff=plural-crypt
.gitattributes !filter !diff
`

const Gitignore = `/**/.terraform
/**/.terraform*
/**/terraform.tfstate*
/bin
*~
.idea
*.swp
*.swo
.DS_STORE
.vscode
`

// IMPORTANT
// Repo cryptography relies on git smudge and clean filters, which pipe a file into stdin and respond with a new version
// of the file from stdout. If we write anything besides the crypto text, it will no longer be decryptable naturally.
func (p *Plural) cryptoCommands() []cli.Command {
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
			Name:   "recover",
			Usage:  "recovers repo encryption keys from a working k8s cluster",
			Action: initKubeconfig(p.handleRecover),
		},
		{
			Name:   "random",
			Usage:  "generates a random string",
			Action: randString,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "len",
					Usage: "the length of the string to generate",
					Value: 32,
				},
			},
		},
		{
			Name:   "ssh-keygen",
			Usage:  "generate an ed5519 keypair for use in git ssh",
			Action: affirmed(handleKeygen, "This command will autogenerate an ed5519 keypair, without passphrase. Sound good?", "PLURAL_CRYPTO_SSH_KEYGEN"),
		},
		{
			Name:   "export",
			Usage:  "dumps the current aes key to stdout",
			Action: exportKey,
		},
		{
			Name:      "share",
			Usage:     "allows a list of plural users to decrypt this repository",
			ArgsUsage: "",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:     "email",
					Usage:    "a email to share with (multiple allowed)",
					Required: true,
				},
			},
			Action: p.handleCryptoShare,
		},
		{
			Name:  "setup-keys",
			Usage: "creates an age keypair, and uploads the public key to plural for use in plural crypto share",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "name",
					Usage:    "a name for the key",
					Required: true,
				},
			},
			Action: p.handleSetupKeys,
		},
		{
			Name:        "backups",
			Usage:       "manages backups of your encryption keys",
			Subcommands: p.backupCommands(),
		},
		{
			Name:   "fingerprint",
			Usage:  "generates a file with the key fingerprint",
			Action: keyFingerprint,
		},
	}
}

func (p *Plural) backupCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "list",
			Usage:  "lists your current key backups",
			Action: p.listBackups,
		},
		{
			Name:   "create",
			Usage:  "creates a backup for your current key",
			Action: affirmed(p.createBackup, backupMsg, "PLURAL_BACKUPS_CREATE"),
		},
		{
			Name:      "restore",
			Usage:     "restores a key backup as your current encryption key",
			ArgsUsage: "NAME",
			Action:    requireArgs(p.restoreBackup, []string{"NAME"}),
		},
	}
}

func handleEncrypt(c *cli.Context) error {
	data, err := io.ReadAll(os.Stdin)
	if bytes.HasPrefix(data, prefix) {
		_, err := os.Stdout.Write(data)
		if err != nil {
			return err
		}
		return nil
	}

	if err != nil {
		return err
	}
	cryptoProv, err := crypto.Build()
	if err != nil {
		return err
	}

	result, err := crypto.Encrypt(cryptoProv, data)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(prefix)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(result)
	if err != nil {
		return err
	}
	return nil
}

func handleDecrypt(c *cli.Context) error {
	var file io.Reader
	if c.Args().Present() {
		p, _ := filepath.Abs(c.Args().First())
		f, err := os.Open(p)
		defer func(f *os.File) {
			_ = f.Close()
		}(f)
		if err != nil {
			return err
		}
		file = f
	} else {
		file = os.Stdin
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	if !bytes.HasPrefix(data, prefix) {
		_, err := os.Stdout.Write(data)
		if err != nil {
			return err
		}
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

	_, err = os.Stdout.Write(result)
	if err != nil {
		return err
	}
	return nil
}

// CheckGitCrypt method checks if the .gitattributes and .gitignore files exist and have desired content.
// Some old repos can have fewer files to encrypt and must be updated.
func CheckGitCrypt(c *cli.Context) error {
	if !utils.Exists(GitAttributesFile) || !utils.Exists(GitIgnoreFile) {
		return cryptoInit(c)
	}
	toCompare := map[string]string{GitAttributesFile: Gitattributes, GitIgnoreFile: Gitignore}

	for file, content := range toCompare {
		equal, err := utils.CompareFileContent(file, content)
		if err != nil {
			return err
		}
		if !equal {
			return cryptoInit(c)
		}
	}

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

	if err := utils.WriteFile(GitAttributesFile, []byte(Gitattributes)); err != nil {
		return err
	}

	if err := utils.WriteFile(GitIgnoreFile, []byte(Gitignore)); err != nil {
		return err
	}

	_, err := crypto.Build()
	return err
}

func (p *Plural) handleCryptoShare(c *cli.Context) error {
	p.InitPluralClient()
	emails := c.StringSlice("email")
	if err := crypto.SetupAge(p.Client, emails); err != nil {
		return err
	}

	prov, err := crypto.BuildAgeProvider()
	if err != nil {
		return err
	}

	return crypto.Flush(prov)
}

func (p *Plural) handleSetupKeys(c *cli.Context) error {
	p.InitPluralClient()
	name := c.String("name")
	if err := crypto.SetupIdentity(p.Client, name); err != nil {
		return err
	}

	utils.Success("Public key uploaded successfully\n")
	return nil
}

func handleUnlock(c *cli.Context) error {
	_, err := crypto.Build()
	if err != nil {
		return err
	}

	repoRoot, err := git.Root()
	if err != nil {
		return err
	}

	gitIndex, _ := filepath.Abs(filepath.Join(repoRoot, ".git", "index"))
	dump, err := os.CreateTemp("", "index.bak")
	if err != nil {
		return err
	}

	if err := os.Rename(gitIndex, dump.Name()); err != nil {
		return err
	}

	if err := gitCommand("checkout", "HEAD", "--", repoRoot).Run(); err != nil {
		_ = os.Rename(dump.Name(), gitIndex)
		return errUnlock
	}

	os.Remove(dump.Name())
	return nil
}

func exportKey(c *cli.Context) error {
	key, err := crypto.Materialize()
	if err != nil {
		return err
	}
	marshal, err := key.Marshal()
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(marshal)
	if err != nil {
		return err
	}
	return nil
}

func importKey(c *cli.Context) error {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	key, err := crypto.Import(data)
	if err != nil {
		return err
	}
	return key.Flush()
}

func randString(c *cli.Context) error {
	var err error
	intVar := c.Int("len")
	len := c.Args().Get(0)
	if len != "" {
		intVar, err = strconv.Atoi(len)
		if err != nil {
			return err
		}
	}
	str, err := crypto.RandStr(intVar)
	if err != nil {
		return err
	}

	fmt.Println(str)
	return nil
}

func handleKeygen(c *cli.Context) error {
	path, err := homedir.Expand("~/.ssh")
	if err != nil {
		return err
	}

	pub, priv, err := scm.GenerateKeys(false)
	if err != nil {
		return err
	}

	filename, ok := utils.GetEnvStringValue("PLURAL_CRYPTO_KEYPAIR_NAME")
	if !ok {
		input := &survey.Input{Message: "What do you want to name your keypair?", Default: "id_plrl"}
		err = survey.AskOne(input, &filename, survey.WithValidator(func(val interface{}) error {
			name, _ := val.(string)
			if utils.Exists(filepath.Join(path, name)) {
				return fmt.Errorf("File ~/.ssh/%s already exists", name)
			}

			return nil
		}))
		if err != nil {
			return err
		}
	}

	if err := os.WriteFile(filepath.Join(path, filename), []byte(priv), 0600); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(path, filename+".pub"), []byte(pub), 0644); err != nil {
		return err
	}

	return nil
}

func (p *Plural) handleRecover(c *cli.Context) error {
	if err := p.InitKube(); err != nil {
		return err
	}

	secret, err := p.Secret("console", "console-conf")
	if err != nil {
		return err
	}

	key, ok := secret.Data["key"]
	if !ok {
		return fmt.Errorf("could not find `key` in console-conf secret")
	}

	aesKey, err := crypto.Import(key)
	if err != nil {
		return err
	}

	if err := crypto.Setup(aesKey.Key); err != nil {
		return err
	}

	utils.Success("Key successfully synced locally!\n")
	fmt.Println("you might need to run `plural crypto init` and `plural crypto setup-keys` to decrypt any repos with your new key")
	return nil
}

func (p *Plural) listBackups(c *cli.Context) error {
	p.InitPluralClient()

	backups, err := p.Client.ListKeyBackups()
	if err != nil {
		return api.GetErrorResponse(err, "ListKeyBackups")
	}

	headers := []string{"Name", "Repositories", "Digest", "Created On"}
	return utils.PrintTable(backups, headers, func(back *api.KeyBackup) ([]string, error) {
		return []string{back.Name, strings.Join(back.Repositories, ", "), back.Digest, back.InsertedAt}, nil
	})
}

func (p *Plural) createBackup(c *cli.Context) error {
	p.InitPluralClient()
	return crypto.BackupKey(p.Client)
}

func (p *Plural) restoreBackup(c *cli.Context) error {
	p.InitPluralClient()
	name := c.Args().First()
	return crypto.DownloadBackup(p.Client, name)
}

func keyFingerprint(_ *cli.Context) error {
	return crypto.CreateKeyFingerprintFile()
}
