package config_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/pluralsh/plural/pkg/config"
	pluraltest "github.com/pluralsh/plural/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestExists(t *testing.T) {
	tests := []struct {
		name             string
		createConfigFile bool
		expectedResponse bool
	}{
		{
			name:             "test when config file doesn't exist",
			expectedResponse: false,
		},
		{
			name:             "test when config file exists",
			expectedResponse: true,
			createConfigFile: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "config")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			os.Setenv("HOME", dir)
			defer os.Unsetenv("HOME")

			if test.createConfigFile {
				defaultConfig := pluraltest.GenDefaultConfig()
				err := defaultConfig.Save(config.ConfigName)
				assert.NoError(t, err)
			}

			result := config.Exists()
			assert.Equal(t, test.expectedResponse, result)
		})
	}
}

func TestRead(t *testing.T) {
	tests := []struct {
		name             string
		createConfigFile bool
		expectedResponse config.Config
	}{
		{
			name:             "test read config file when file doesn't exist",
			expectedResponse: config.Config{},
		},
		{
			name:             "test read config file when file exists",
			createConfigFile: true,
			expectedResponse: pluraltest.GenDefaultConfig(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "config")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			os.Setenv("HOME", dir)
			defer os.Unsetenv("HOME")

			if test.createConfigFile {
				defaultConfig := pluraltest.GenDefaultConfig()
				err := defaultConfig.Save(config.ConfigName)
				assert.NoError(t, err)
			}

			result := config.Read()
			assert.Equal(t, test.expectedResponse, result)
		})
	}
}

func TestProfiles(t *testing.T) {
	defaultConfig := pluraltest.GenDefaultConfig()
	tests := []struct {
		name             string
		createConfigFile bool
		expectedResponse []*config.VersionedConfig
	}{
		{
			name:             "test when profile config file doesn't exist",
			expectedResponse: []*config.VersionedConfig{},
		},
		{
			name:             "test when profile config file exists",
			createConfigFile: true,
			expectedResponse: []*config.VersionedConfig{
				{
					ApiVersion: "platform.plural.sh/v1alpha1",
					Kind:       "Config",
					Spec:       &defaultConfig,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "config")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			os.Setenv("HOME", dir)
			defer os.Unsetenv("HOME")

			// create config in order to init plural directory
			defaultConfig := pluraltest.GenDefaultConfig()
			err = defaultConfig.Save(config.ConfigName)
			assert.NoError(t, err)

			if test.createConfigFile {
				defaultConfig := pluraltest.GenDefaultConfig()
				err := defaultConfig.Save("profile.yml")
				assert.NoError(t, err)
			}

			results, err := config.Profiles()
			assert.NoError(t, err)
			assert.Equal(t, test.expectedResponse, results)
		})
	}
}

func TestAmend(t *testing.T) {
	tests := []struct {
		name             string
		createConfigFile bool
		key              string
		value            string
		expectedError    bool
		expectedConfig   config.Config
	}{
		{
			name:          "test amend config file when file doesn't exist",
			key:           "token",
			value:         "cdf",
			expectedError: false,
			expectedConfig: config.Config{
				Token: "cdf",
			},
		},
		{
			name:             "test amend token",
			createConfigFile: true,
			key:              "token",
			value:            "cdf",
			expectedError:    false,
			expectedConfig: func() config.Config {
				defaultConf := pluraltest.GenDefaultConfig()
				defaultConf.Token = "cdf"
				return defaultConf
			}(),
		},
		{
			name:          "test amend config file with wrong key",
			key:           "unknown",
			value:         "abc",
			expectedError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "config")
			assert.NoError(t, err)
			defer os.RemoveAll(dir)

			os.Setenv("HOME", dir)
			defer os.Unsetenv("HOME")

			if test.createConfigFile {
				defaultConfig := pluraltest.GenDefaultConfig()
				err := defaultConfig.Save(config.ConfigName)
				assert.NoError(t, err)
			}

			err = config.Amend(test.key, test.value)
			if test.expectedError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			result := config.Read()
			assert.Equal(t, test.expectedConfig, result)
		})
	}
}
