package main_test

import (
	"os"
	"testing"

	plural "github.com/pluralsh/plural/cmd/plural"
	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/test/mocks"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		me            *api.Me
		expectedError string
	}{
		{
			name: `test init when demo cluster is running`,
			args: []string{plural.ApplicationName, "init"},
			me: &api.Me{
				Demoing: true,
			},
			expectedError: plural.DemoingErrorMsg,
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

			client := mocks.NewClient(t)
			client.On("Me").Return(test.me, nil)
			app := plural.CreateNewApp(&plural.Plural{Client: client})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			_, err = captureStdout(app, os.Args)
			if test.expectedError != "" {
				assert.Error(t, err)
				assert.Equal(t, test.expectedError, err.Error())
			}
		})
	}
}
