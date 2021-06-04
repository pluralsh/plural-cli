package crypto

import (
	"fmt"
	"os"
	"path/filepath"
	"io/ioutil"
	"io"
	"bytes"
	"strings"
	"time"
	"filippo.io/age"
	"gopkg.in/yaml.v2"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/api"
)

const (
	identityFile = "identities.yml"
	gitignore = "identity\n"
)

type AgeProvider struct {
	Identity *age.X25519Identity
	Key *AESKey
}

type Age struct {
	RepoKey string
	Identities []*AgeIdentity
}

type AgeIdentity struct {
	Key   string
	Email string
}

func (prov *AgeProvider) SymmetricKey() ([]byte, error) {
	dummy := &KeyProvider{key: prov.Key.Key}
	return dummy.SymmetricKey()
}

func (prov *AgeProvider) ID() string {
	dummy := &KeyProvider{key: prov.Key.Key}
	return dummy.ID()
}

func (prov *AgeProvider) Marshall() ([]byte, error) {
	conf := Config{
		Version: "crypto.plural.sh/v1",
		Type: AGE,
		Id: prov.ID(),
		Context: map[string]interface{}{},
	}

	return yaml.Marshal(conf)
}

func (prov *AgeProvider) decrypt(content []byte) ([]byte, error) {
	buf := bytes.NewBuffer(content)
	ident, err := Identity()
	reader, err := age.Decrypt(buf, ident)
	if err != nil {
		return []byte{}, err
	}

	var out bytes.Buffer
	_, err = io.Copy(&out, reader)
	return out.Bytes(), err
}

func BuildAgeProvider() (prov *AgeProvider, err error) {
	ident, err := Identity()
	if err != nil {
		return
	}

	prov = &AgeProvider{Identity: ident}
	contents, err := ioutil.ReadFile(filepath.Join(cryptPath(), "key"))
	if err != nil {
		return
	}

	keycontent, err := prov.decrypt(contents)
	if err != nil {
		return
	}
	aes, err := DeserializeKey(keycontent)
	prov.Key = aes
	return
}

func Identity() (*age.X25519Identity, error) {
	return generateIdentity(getAgePath())
}

func SetupAge(emails []string) error {
	client := api.NewClient()
	age, err := setupAgeConfig()
	if err != nil {
		return err
	}

	// if any additional emails were specified, add them now
	if len(emails) > 0 {
		keys, err := client.ListKeys(emails)
		if err != nil {
			return err
		}

		idents := age.Identities
		for _, key := range keys {
			idents = append(idents, &AgeIdentity{Key: key.Content, Email: key.User.Email})
		}

		age.Identities = idents
	}

	keyPath := filepath.Join(cryptPath(), "key")
	// repo key already exists, so re-encrypt using new age config
	if utils.Exists(keyPath) {
		prov, err := BuildAgeProvider()
		if err != nil {
			return err
		}

		keydata, _ := prov.Key.Marshal()
		return age.WriteKeyFile(keyPath, keydata)
	}

	key, _ := Materialize()
	keydata, err := key.Marshal()
	if err != nil {
		return err
	}

	return age.WriteKeyFile(keyPath, keydata)
}

func (a *Age) Recipients() []age.Recipient {
	recipients := make([]age.Recipient, 0)

	for _, ident := range a.Identities {
		r, err := age.ParseX25519Recipient(ident.Key)
		if err != nil {
			panic(err)
		}
		recipients = append(recipients, r)
	}

	r, err := age.ParseX25519Recipient(a.RepoKey)
	if err != nil {
		panic(err)
	}
	return append(recipients, r)
}

func (a *Age) encrypt(content []byte) ([]byte, error) {
	var buf bytes.Buffer
	recips := a.Recipients()
	writer, err := age.Encrypt(&buf, recips...)
	if err != nil {
		return buf.Bytes(), err
	}

	writer.Write(content)
	writer.Close()
	return buf.Bytes(), nil
}

func (age *Age) WriteKeyFile(path string, keydata []byte) error {
	encrypted, err := age.encrypt(keydata)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(path, encrypted, 0644); err != nil {
		return err
	}
	// always flush current age config after writing key to preserve state
	return age.Flush()
}

func (age *Age) Flush() error {
	contents, err := yaml.Marshal(age)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(cryptPath(), identityFile), contents, 0644)
}

func SetupIdentity(name string) error {
	client := api.NewClient()
	userIdentity, err := generateIdentity(getAgePath())
	if err != nil {
		return err
	}
	
	return client.CreateKey(name, userIdentity.Recipient().String())
}

func setupAgeConfig() (*Age, error) {
	base := cryptPath()

	// first set up directory and gitignore files
	if err := os.MkdirAll(base, os.ModePerm); err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(filepath.Join(base, ".gitignore"), []byte(gitignore), 0644); err != nil {
		return nil, err
	}

	// ensure a repo identity is present for use in console deployments (primarily)
	userIdentity, err := generateIdentity(getAgePath())
	if err != nil {
		return nil, err
	}

	repoIdentity, err := generateIdentity(filepath.Join(base, "identity"))
	if err != nil {
		return nil, err
	}

	// create the 
	conf := config.Read()
	age := &Age{
		RepoKey: repoIdentity.Recipient().String(), 
		Identities: []*AgeIdentity{
			{Email: conf.Email, Key: userIdentity.Recipient().String()},
		},
	}
	return age, nil
}

func identityFromString(contents string) (*age.X25519Identity, error) {
	for _, line := range strings.Split(string(contents), "\n") {
		if strings.HasPrefix(line, "#") || line == "" { continue }
		return age.ParseX25519Identity(line)
	}

	return nil, fmt.Errorf("No identity found")
}

func generateIdentity(path string) (*age.X25519Identity, error) {
	if utils.Exists(path) {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}

		return identityFromString(string(contents))
	}

	k, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fmt.Fprintf(f, "# created: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(f, "# public key: %s\n", k.Recipient())
	fmt.Fprintf(f, "%s\n", k)
	return k, nil
}

func getAgePath() string {
	folder, _ := os.UserHomeDir()
	return filepath.Join(folder, ".plural", "identity")
}

func cryptPath() string {
	root, _ := utils.RepoRoot()
	return filepath.Join(root, ".plural-crypt")
}