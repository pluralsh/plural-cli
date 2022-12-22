package crypto_test

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/crypto"
	"github.com/pluralsh/plural/pkg/test/mocks"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/pluralsh/plural/pkg/utils/pathing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v2"
)

func TestSetupAge(t *testing.T) {
	tests := []struct {
		name          string
		expectedError string
		expected      []*crypto.AgeIdentity
		emails        []string
		keys          []*api.PublicKey
	}{
		{
			name:   `when one of the user has no keys setup`,
			emails: []string{"test@plural.sh", "test-1@plural.sh"},
			keys: []*api.PublicKey{
				{
					Id:      "1",
					Content: "abc",
					User: &api.User{
						Id:    "1",
						Email: "test@plural.sh",
						Name:  "test",
					},
				},
			},
			expectedError: "Some of the users [test-1@plural.sh] have no keys setup",
		},
		{
			name:          `when all users have no keys setup`,
			emails:        []string{"test@plural.sh", "test-1@plural.sh"},
			keys:          []*api.PublicKey{},
			expectedError: "Some of the users [test@plural.sh test-1@plural.sh] have no keys setup",
		},
		{
			name:   `append user to identities.yaml`,
			emails: []string{"test@plural.sh"},
			keys: []*api.PublicKey{
				{
					Id:      "1",
					Content: "age1wqc2hk954ukemelys5gxdwlqve8ev0e88hvl3cjhfcvq65gwgvsqkmq9dn",
					User: &api.User{
						Id:    "1",
						Email: "test@plural.sh",
						Name:  "test",
					},
				},
			},
			expected: []*crypto.AgeIdentity{
				{
					Key:   "age1wqc2hk954ukemelys5gxdwlqve8ev0e88hvl3cjhfcvq65gwgvsqkmq9dn",
					Email: "test@plural.sh",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir, err := os.MkdirTemp("", "config")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			os.Setenv("HOME", dir)
			defer os.Unsetenv("HOME")

			err = os.Chdir(dir)
			assert.NoError(t, err)
			_, err = git.Init()
			assert.NoError(t, err)

			err = os.WriteFile(path.Join(dir, "crypto.yml"), []byte("abc"), 0644)
			assert.NoError(t, err)

			err = os.MkdirAll(path.Join(dir, ".plural"), os.ModePerm)
			assert.NoError(t, err)
			err = os.WriteFile(path.Join(dir, ".plural", "key"), []byte("key: abc"), 0644)
			assert.NoError(t, err)

			err = os.MkdirAll(path.Join(dir, ".plural-crypt"), os.ModePerm)
			assert.NoError(t, err)
			identities := `repokey: age19lm6v2l4czn0rlhr7xy3g7ek8w4z7dn9qvestez8n5x6yxrs3v2semmafe
identities: []
`
			err = os.WriteFile(path.Join(dir, ".plural-crypt", "identities.yml"), []byte(identities), 0644)
			assert.NoError(t, err)

			client := mocks.NewClient(t)
			client.On("ListKeys", mock.Anything).Return(test.keys, nil)

			err = crypto.SetupAge(client, test.emails)
			if test.expectedError != "" {
				assert.Equal(t, test.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
				path := pathing.SanitizeFilepath(filepath.Join(dir, ".plural-crypt", "identities.yml"))
				age := &crypto.Age{}
				contents, err := os.ReadFile(path)
				assert.NoError(t, err)

				err = yaml.Unmarshal(contents, age)
				assert.NoError(t, err)
				assert.Equal(t, age.Identities, test.expected)
			}
		})
	}
}
