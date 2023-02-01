package main_test

import (
	"encoding/base64"
	"os"
	"path"
	"testing"

	"github.com/pluralsh/plural/pkg/utils"
	"github.com/urfave/cli"

	plural "github.com/pluralsh/plural/cmd/plural"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/config"
	pluraltest "github.com/pluralsh/plural/pkg/test"
	"github.com/pluralsh/plural/pkg/test/mocks"
	"github.com/pluralsh/plural/pkg/utils/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
)

func TestSetupKeys(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError string
	}{
		{
			name:          `test "crypto setup-keys" without name flag`,
			args:          []string{plural.ApplicationName, "crypto", "setup-keys"},
			expectedError: "Required flag \"name\" not set",
		},
		{
			name: `test "crypto setup-keys"`,
			args: []string{plural.ApplicationName, "crypto", "setup-keys", "--name", "test"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create temp environment
			currentDir, err := os.Getwd()
			assert.NoError(t, err)
			dir, err := os.MkdirTemp("", "config")
			assert.NoError(t, err)
			defer func(path, currentDir string) {
				_ = os.RemoveAll(path)
				_ = os.Chdir(currentDir)
			}(dir, currentDir)
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
			name:          `test "crypto share" without name flag`,
			args:          []string{plural.ApplicationName, "crypto", "share"},
			expectedError: "Required flag \"email\" not set",
		},
		{
			name: `test "crypto share"`,
			args: []string{plural.ApplicationName, "crypto", "share", "--email", "test@email.com"},
			keys: []*api.PublicKey{
				{
					Id:      "abc",
					Content: "age1wqc2hk954ukemelys5gxdwlqve8ev0e88hvl3cjhfcvq65gwgvsqkmq9dn",
					User: &api.User{
						Email: "test@email.com",
					},
				},
			},
		},
		{
			name: `test "crypto share" where test user has no key setup`,
			args: []string{plural.ApplicationName, "crypto", "share", "--email", "test@email.com"},
			keys: []*api.PublicKey{
				{
					Id:      "abc",
					Content: "age1wqc2hk954ukemelys5gxdwlqve8ev0e88hvl3cjhfcvq65gwgvsqkmq9dn",
					User: &api.User{
						Email: "example@email.com",
					},
				},
			},
			expectedError: "Some of the users [test@email.com] have no keys setup",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			currentDir, err := os.Getwd()
			assert.NoError(t, err)
			dir, err := os.MkdirTemp("", "config")
			assert.NoError(t, err)
			defer func(path, currentDir string) {
				_ = os.RemoveAll(path)
				_ = os.Chdir(currentDir)
			}(dir, currentDir)
			err = os.Chdir(dir)
			assert.NoError(t, err)
			os.Setenv("HOME", dir)
			defer os.Unsetenv("HOME")
			defaultConfig := pluraltest.GenDefaultConfig()
			err = defaultConfig.Save(config.ConfigName)
			assert.NoError(t, err)

			client := mocks.NewClient(t)
			if test.keys != nil {
				client.On("ListKeys", mock.Anything).Return(test.keys, nil)
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
			currentDir, err := os.Getwd()
			assert.NoError(t, err)
			dir, err := os.MkdirTemp("", "config")
			assert.NoError(t, err)
			defer func(path, currentDir string) {
				_ = os.RemoveAll(path)
				_ = os.Chdir(currentDir)
			}(dir, currentDir)

			os.Setenv("HOME", dir)
			defer os.Unsetenv("HOME")

			defaultConfig := pluraltest.GenDefaultConfig()
			err = defaultConfig.Save(config.ConfigName)
			assert.NoError(t, err)

			client := mocks.NewClient(t)
			kube := mocks.NewKube(t)

			if test.keyContent != "" {
				err := os.WriteFile(path.Join(dir, ".plural", "key"), []byte(test.keyContent), 0644)
				assert.NoError(t, err)
			}

			kube.On("Secret", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(test.secret, nil)

			app := plural.CreateNewApp(&plural.Plural{Client: client, Kube: kube})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			_, err = captureStdout(app, os.Args)
			assert.NoError(t, err)

			b, err := os.ReadFile(path.Join(dir, ".plural", "key"))
			assert.NoError(t, err)
			assert.Equal(t, test.expectedKey, string(b))
		})
	}
}

func TestCheckGitCrypt(t *testing.T) {
	tests := []struct {
		name        string
		createFiles bool
	}{
		{
			name: "test when .gitattributes and .gitignore don't exist",
		},
		{
			name:        "test when .gitattributes and .gitignore exist with the wrong content",
			createFiles: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create temp environment
			currentDir, err := os.Getwd()
			assert.NoError(t, err)
			dir, err := os.MkdirTemp("", "config")
			assert.NoError(t, err)
			defer func(path, currentDir string) {
				_ = os.RemoveAll(path)
				_ = os.Chdir(currentDir)
			}(dir, currentDir)
			os.Setenv("HOME", dir)
			defer os.Unsetenv("HOME")
			defaultConfig := pluraltest.GenDefaultConfig()
			err = defaultConfig.Save(config.ConfigName)
			assert.NoError(t, err)
			err = os.WriteFile(path.Join(dir, ".plural", "key"), []byte("key: abc"), 0644)
			assert.NoError(t, err)

			err = os.Chdir(dir)
			assert.NoError(t, err)
			_, err = git.Init()
			assert.NoError(t, err)

			gitAttributes := path.Join(dir, plural.GitAttributesFile)
			gitIgnore := path.Join(dir, plural.GitIgnoreFile)

			if test.createFiles {
				err = utils.WriteFile(gitIgnore, []byte(plural.Gitignore+"some extra"))
				assert.NoError(t, err)
				err = utils.WriteFile(gitAttributes, []byte(plural.Gitattributes+"abc"))
				assert.NoError(t, err)
			}

			// test CheckGitCrypt
			err = plural.CheckGitCrypt(&cli.Context{})
			assert.NoError(t, err)

			// the files should exist
			assert.True(t, utils.Exists(gitAttributes), ".gitattributes should exist")
			assert.True(t, utils.Exists(gitIgnore), ".gitignore should exist")

			attributes, err := utils.ReadFile(gitAttributes)
			assert.NoError(t, err)
			assert.Equal(t, attributes, plural.Gitattributes)

			ignore, err := utils.ReadFile(gitIgnore)
			assert.NoError(t, err)
			assert.Equal(t, ignore, plural.Gitignore)
		})
	}
}
