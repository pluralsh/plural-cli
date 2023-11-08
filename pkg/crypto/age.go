package crypto

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"filippo.io/age"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/utils"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/pluralsh/polly/algorithms"
	"github.com/pluralsh/polly/containers"
	"github.com/samber/lo"
	"gopkg.in/yaml.v2"
)

const (
	identityFile = "identities.yml"
	gitignore    = "identity\n"
)

type AgeProvider struct {
	Identity *age.X25519Identity
	Key      *AESKey
}

type Age struct {
	RepoKey    string
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
		Type:    AGE,
		Id:      prov.ID(),
	}

	return yaml.Marshal(conf)
}

func (prov *AgeProvider) decrypt(content []byte) ([]byte, error) {
	buf := bytes.NewBuffer(content)
	ident, err := Identity()
	if err != nil {
		return []byte{}, err
	}
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
	contents, err := os.ReadFile(pathing.SanitizeFilepath(filepath.Join(cryptPath(), "key")))
	if err != nil {
		return
	}

	keycontent, err := prov.decrypt(contents)
	if err != nil {
		return
	}
	aes, err := DeserializeKey(keycontent)
	if err != nil {
		return
	}
	prov.Key = aes
	return
}

func Identity() (*age.X25519Identity, error) {
	return generateIdentity(getAgePath())
}

func findMissingKeyForEmail(emails []string, keys []*api.PublicKey) []string {
	if len(emails) == 0 {
		return []string{}
	}
	if len(keys) == 0 {
		// in case of empty or nil objects
		return emails
	}
	emailSet := containers.ToSet[string](emails)
	algorithms.Map(keys, func(key *api.PublicKey) string {
		if key.User != nil && key.User.Email != "" {
			emailSet.Remove(key.User.Email)
		}
		return ""
	})

	return emailSet.List()
}

func SetupAge(client api.Client, emails []string) error {
	ageConfig, err := setupAgeConfig()
	if err != nil {
		return err
	}

	// if any additional emails were specified, add them now
	if len(emails) > 0 {
		keys, err := client.ListKeys(emails)
		if err != nil {
			return api.GetErrorResponse(err, "ListKeys")
		}

		missingEmails := findMissingKeyForEmail(emails, keys)
		if len(missingEmails) > 0 {
			return fmt.Errorf("Some of the users %v have no keys setup", missingEmails)
		}

		idents := algorithms.Map(keys, func(key *api.PublicKey) *AgeIdentity { return &AgeIdentity{Key: key.Content, Email: key.User.Email} })
		ageConfig.Identities = lo.UniqBy(append(idents, ageConfig.Identities...), func(id *AgeIdentity) string { return fmt.Sprintf("%s::%s", id.Email, id.Key) })
	}

	keyPath := pathing.SanitizeFilepath(filepath.Join(cryptPath(), "key"))
	// repo key already exists, so re-encrypt using new age config
	if utils.Exists(keyPath) {
		prov, err := BuildAgeProvider()
		if err != nil {
			return err
		}

		keydata, _ := prov.Key.Marshal()
		return ageConfig.WriteKeyFile(keyPath, keydata)
	}

	key, _ := Materialize()
	keydata, err := key.Marshal()
	if err != nil {
		return err
	}

	return ageConfig.WriteKeyFile(keyPath, keydata)
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

	if _, err := writer.Write(content); err != nil {
		return buf.Bytes(), err
	}
	if err := writer.Close(); err != nil {
		return buf.Bytes(), err
	}
	return buf.Bytes(), nil
}

func (age *Age) WriteKeyFile(path string, keydata []byte) error {
	encrypted, err := age.encrypt(keydata)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, encrypted, 0644); err != nil {
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

	return os.WriteFile(pathing.SanitizeFilepath(filepath.Join(cryptPath(), identityFile)), contents, 0644)
}

func SetupIdentity(client api.Client, name string) error {
	userIdentity, err := generateIdentity(getAgePath())
	if err != nil {
		return err
	}

	if err := client.CreateKey(name, userIdentity.Recipient().String()); err != nil {
		return api.GetErrorResponse(err, "CreateKey")
	}
	return nil
}

func setupAgeConfig() (*Age, error) {
	base := cryptPath()
	path := pathing.SanitizeFilepath(filepath.Join(base, "identities.yml"))

	if utils.Exists(path) {
		age := &Age{}
		contents, err := os.ReadFile(path)
		if err != nil {
			return age, err
		}

		err = yaml.Unmarshal(contents, age)
		return age, err
	}

	// first set up directory and gitignore files
	if err := os.MkdirAll(base, os.ModePerm); err != nil {
		return nil, err
	}

	if err := os.WriteFile(pathing.SanitizeFilepath(filepath.Join(base, ".gitignore")), []byte(gitignore), 0644); err != nil {
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
	ageOutput := &Age{
		RepoKey: repoIdentity.Recipient().String(),
		Identities: []*AgeIdentity{
			{Email: conf.Email, Key: userIdentity.Recipient().String()},
		},
	}
	return ageOutput, nil
}

func identityFromString(contents string) (*age.X25519Identity, error) {
	for _, line := range strings.Split(string(contents), "\n") {
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		return age.ParseX25519Identity(line)
	}

	return nil, fmt.Errorf("No identity found")
}

func generateIdentity(path string) (*age.X25519Identity, error) {
	if utils.Exists(path) {
		contents, err := os.ReadFile(path)
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
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	if _, err := fmt.Fprintf(f, "# created: %s\n", time.Now().Format(time.RFC3339)); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(f, "# public key: %s\n", k.Recipient()); err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(f, "%s\n", k); err != nil {
		return nil, err
	}
	return k, nil
}

func getAgePath() string {
	folder, _ := os.UserHomeDir()
	return pathing.SanitizeFilepath(filepath.Join(folder, ".plural", "identity"))
}

func cryptPath() string {
	root, _ := git.Root()
	return pathing.SanitizeFilepath(filepath.Join(root, ".plural-crypt"))
}
