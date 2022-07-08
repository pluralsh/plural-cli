package test

import (
	"github.com/pluralsh/plural/pkg/config"
)

func GenDefaultConfig() config.Config {
	return config.Config{
		Email:           "test@plural.sh",
		Token:           "abc",
		NamespacePrefix: "test",
		Endpoint:        "http://example.com",
		LockProfile:     "abc",
		ReportErrors:    false,
	}
}
