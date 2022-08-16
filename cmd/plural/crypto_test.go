package main_test

import (
	"encoding/base64"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/pluralsh/plural/pkg/config"
	pluraltest "github.com/pluralsh/plural/pkg/test"
	"github.com/stretchr/testify/mock"

	plural "github.com/pluralsh/plural/cmd/plural"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/test/mocks"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestSetupKeys(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError string
	}{
		{
			name:          `test "crypto setup-keys" without name argument`,
			args:          []string{plural.ApplicationName, "crypto", "setup-keys"},
			expectedError: "Not enough arguments provided: needs name. Try running --help to see usage.",
		},
		{
			name: `test "crypto setup-keys"`,
			args: []string{plural.ApplicationName, "crypto", "setup-keys", "test"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create temp environment
			dir, err := ioutil.TempDir("", "config")
			assert.NoError(t, err)
			defer func(path string) {
				_ = os.RemoveAll(path)
			}(dir)
			os.Setenv("HOME", dir)
			defer os.Unsetenv("HOME")
			defaultConfig := pluraltest.GenDefaultConfig()
			err = defaultConfig.Save(config.ConfigName)
			assert.NoError(t, err)

			client := mocks.NewClient(t)
			if test.expectedError == "" {
				client.On("CreateKey", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
			}
			app := plural.CreateNewApp(&plural.Plural{Client: client})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			_, err = captureStdout(app, os.Args)
			if test.expectedError != "" {
				assert.Equal(t, err.Error(), test.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRandom(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectedLen int
	}{
		{
			name:        `test "crypto random" without len argument, gets default`,
			args:        []string{plural.ApplicationName, "crypto", "random"},
			expectedLen: 32,
		},
		{
			name:        `test "crypto setup-keys"`,
			args:        []string{plural.ApplicationName, "crypto", "random", "10"},
			expectedLen: 10,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewClient(t)
			app := plural.CreateNewApp(&plural.Plural{Client: client})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			res, err := captureStdout(app, os.Args)
			assert.NoError(t, err)
			b, err := base64.StdEncoding.DecodeString(res)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedLen, len(b))
		})
	}
}

func TestShare(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError string
		keys          []*api.PublicKey
	}{
		{
			name:          `test "crypto share" without name argument`,
			args:          []string{plural.ApplicationName, "crypto", "share"},
			expectedError: "Not enough arguments provided: needs email. Try running --help to see usage.",
		},
		{
			name: `test "crypto share"`,
			args: []string{plural.ApplicationName, "crypto", "share", "test@email.com"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create temp environment
			dir, err := ioutil.TempDir("", "config")
			assert.NoError(t, err)
			defer func(path string) {
				_ = os.RemoveAll(path)
			}(dir)
			err = os.Chdir(dir)
			assert.NoError(t, err)
			defaultConfig := pluraltest.GenDefaultConfig()
			err = defaultConfig.Save(config.ConfigName)
			assert.NoError(t, err)

			client := mocks.NewClient(t)
			if test.expectedError == "" {
				client.On("ListKeys", mock.Anything).Return(nil, nil)
			}
			app := plural.CreateNewApp(&plural.Plural{Client: client})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			_, err = captureStdout(app, os.Args)
			if test.expectedError != "" {
				assert.Equal(t, err.Error(), test.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRecover(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		secret      *v1.Secret
		keyContent  string
		expectedKey string
	}{
		{
			name: `test "crypto recover" when key file doesn't exist`,
			args: []string{plural.ApplicationName, "crypto", "recover"},
			secret: &v1.Secret{
				Data: map[string][]byte{
					"key": []byte("key: gKNJBnflqQA6lfUKLWMwl7CMJk4j+qqG9jnGYdTvwTk="),
				},
			},
			expectedKey: "key: gKNJBnflqQA6lfUKLWMwl7CMJk4j+qqG9jnGYdTvwTk=\n",
		},
		{
			name: `test "crypto recover" when key file is broken`,
			args: []string{plural.ApplicationName, "crypto", "recover"},
			secret: &v1.Secret{
				Data: map[string][]byte{
					"key": []byte("key: gKNJBnflqQA6lfUKLWMwl7CMJk4j+qqG9jnGYdTvwTk="),
				},
			},
			keyContent:  "      key: |\n        key: |\n          key: abc",
			expectedKey: "key: gKNJBnflqQA6lfUKLWMwl7CMJk4j+qqG9jnGYdTvwTk=\n",
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

			defaultConfig := pluraltest.GenDefaultConfig()
			err = defaultConfig.Save(config.ConfigName)
			assert.NoError(t, err)

			client := mocks.NewClient(t)
			kube := mocks.NewKube(t)

			if test.keyContent != "" {
				err := ioutil.WriteFile(path.Join(dir, ".plural", "key"), []byte(test.keyContent), 0644)
				assert.NoError(t, err)
			}

			kube.On("Secret", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(test.secret, nil)

			app := plural.CreateNewApp(&plural.Plural{Client: client, Kube: kube})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			_, err = captureStdout(app, os.Args)
			assert.NoError(t, err)

			b, err := os.ReadFile(dir + "/.plural/key")
			assert.NoError(t, err)
			assert.Equal(t, test.expectedKey, string(b))
		})
	}
}
