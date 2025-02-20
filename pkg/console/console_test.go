package console_test

import (
	"testing"

	"github.com/pluralsh/plural-cli/pkg/console"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeUrl(t *testing.T) {
	tests := []struct {
		name                string
		url, expectedResult string
		expectError         bool
	}{
		{
			name:           `valid url should stay the same`,
			url:            "https://console.test.onplural.sh/gql",
			expectedResult: "https://console.test.onplural.sh/gql",
		},
		{
			name:           `trailing slash should be removed`,
			url:            "https://console.test.onplural.sh/gql/",
			expectedResult: "https://console.test.onplural.sh/gql",
		},
		{
			name:           `protocol should be set to https`,
			url:            "http://console.test.onplural.sh/gql",
			expectedResult: "https://console.test.onplural.sh/gql",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedResult, console.NormalizeUrl(test.url))
		})
	}
}

func TestNormalizeExtUrl(t *testing.T) {
	tests := []struct {
		name                string
		url, expectedResult string
		expectError         bool
	}{
		{
			name:           `valid url should stay the same`,
			url:            "https://console.test.onplural.sh/ext/gql",
			expectedResult: "https://console.test.onplural.sh/ext/gql",
		},
		{
			name:           `trailing slash should be removed`,
			url:            "https://console.test.onplural.sh/ext/gql/",
			expectedResult: "https://console.test.onplural.sh/ext/gql",
		},
		{
			name:           `protocol should be set to https`,
			url:            "http://console.test.onplural.sh/ext/gql",
			expectedResult: "https://console.test.onplural.sh/ext/gql",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedResult, console.NormalizeExtUrl(test.url))
		})
	}
}
