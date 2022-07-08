package main_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"

	plural "github.com/pluralsh/plural/cmd/plural"
	"github.com/pluralsh/plural/pkg/config"
	pluraltest "github.com/pluralsh/plural/pkg/test"
)

func TestPluralConfigCommand(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		createConfig     bool
		expectedResponse string
	}{
		{
			name:             `test "config read" command when config file doesn't exists'`,
			args:             []string{plural.ApplicationName, "config", "read"},
			expectedResponse: "apiVersion: platform.plural.sh/v1alpha1\nkind: Config\nmetadata: null\nspec:\n  email: \"\"\n  token: \"\"\n  namespacePrefix: \"\"\n  endpoint: \"\"\n  lockProfile: \"\"\n  reportErrors: false\n",
		},
		{
			name:             `test "config read" command with default test config'`,
			args:             []string{plural.ApplicationName, "config", "read"},
			createConfig:     true,
			expectedResponse: "apiVersion: platform.plural.sh/v1alpha1\nkind: Config\nmetadata: null\nspec:\n  email: test@plural.sh\n  token: abc\n  namespacePrefix: test\n  endpoint: http://example.com\n  lockProfile: abc\n  reportErrors: false\n",
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

			if test.createConfig {
				defaultConfig := pluraltest.GenDefaultConfig()
				err := defaultConfig.Save(config.ConfigName)
				assert.NoError(t, err)
			}

			app := plural.CreateNewApp()
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			res, err := captureStdout(app, os.Args)
			assert.NoError(t, err)

			assert.Equal(t, test.expectedResponse, res)
		})
	}
}

func captureStdout(app *cli.App, arg []string) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := app.Run(arg)
	if err != nil {
		return "", err
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return "", err
	}
	return buf.String(), nil
}
