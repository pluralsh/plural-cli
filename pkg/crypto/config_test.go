package crypto_test

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/pluralsh/plural/pkg/crypto"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/stretchr/testify/assert"
)

func TestBuild(t *testing.T) {
	tests := []struct {
		name          string
		expectedError string
		expected      string
		keyContent    string
		genConfig     bool
	}{
		{
			name:       `when faulty config exists create default fallbackProvider`,
			keyContent: "key: abc",
			expected:   "SHA256:ungWv48Bz+pBQUDeXa4iI7ADYaOWF3qctBD/YfIAFa0=",
			genConfig:  true,
		},
		{
			name:       `when config doesn't exist create default fallbackProvider`,
			genConfig:  false,
			keyContent: "key: abc",
			expected:   "SHA256:ungWv48Bz+pBQUDeXa4iI7ADYaOWF3qctBD/YfIAFa0=",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create temp environment
			dir, err := ioutil.TempDir("", "config")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			os.Setenv("HOME", dir)
			defer os.Unsetenv("HOME")

			err = os.Chdir(dir)
			assert.NoError(t, err)
			_, err = git.Init()
			assert.NoError(t, err)

			if test.genConfig {
				err := ioutil.WriteFile(path.Join(dir, "crypto.yml"), []byte("abc"), 0644)
				assert.NoError(t, err)
			}

			err = os.MkdirAll(path.Join(dir, ".plural"), os.ModePerm)
			assert.NoError(t, err)
			err = ioutil.WriteFile(path.Join(dir, ".plural", "key"), []byte(test.keyContent), 0644)
			assert.NoError(t, err)

			provider, err := crypto.Build()
			if test.expectedError != "" {
				assert.Equal(t, test.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, provider.ID())
			}
		})
	}
}
