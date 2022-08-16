package main_test

import (
	"os"
	"testing"

	"github.com/pluralsh/plural-operator/api/platform/v1alpha1"
	plural "github.com/pluralsh/plural/cmd/plural"
	"github.com/pluralsh/plural/pkg/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLogsList(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedError string
	}{
		{
			name:          `test "logs list" without repo name`,
			args:          []string{plural.ApplicationName, "logs", "list"},
			expectedError: "Not enough arguments provided: needs REPO. Try running --help to see usage.",
		},
		{
			name: `test "logs list" with repo name`,
			args: []string{plural.ApplicationName, "logs", "list", "test"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := mocks.NewClient(t)
			kube := mocks.NewKube(t)
			if test.expectedError == "" {
				kube.On("LogTailList", mock.AnythingOfType("string")).Return(&v1alpha1.LogTailList{Items: []v1alpha1.LogTail{}}, nil)
			}
			app := plural.CreateNewApp(&plural.Plural{Client: client, Kube: kube})
			app.HelpName = plural.ApplicationName
			os.Args = test.args
			_, err := captureStdout(app, os.Args)
			if test.expectedError != "" {
				assert.Equal(t, test.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
				kube.AssertCalled(t, "LogTailList", mock.AnythingOfType("string"))
			}

		})
	}
}
