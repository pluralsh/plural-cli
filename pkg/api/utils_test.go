package api_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pluralsh/plural-cli/pkg/api"
)

func TestNormalizeProvider(t *testing.T) {
	tests := []struct {
		provider string
		expected string
	}{
		{provider: `aws`, expected: `aws`},
		{provider: `gcp`, expected: `gcp`},
		{provider: `google`, expected: `gcp`},
		{provider: `azure`, expected: `azure`},
	}
	for _, test := range tests {
		t.Run(test.provider, func(t *testing.T) {
			result := api.NormalizeProvider(test.provider)
			assert.Equal(t, result, test.expected)
		})
	}
}

func TestToGQLClientProvider(t *testing.T) {
	tests := []struct {
		provider string
		expected string
	}{
		{provider: `aws`, expected: `AWS`},
		{provider: `gcp`, expected: `GCP`},
		{provider: `google`, expected: `GCP`},
		{provider: `azure`, expected: `AZURE`},
	}
	for _, test := range tests {
		t.Run(test.provider, func(t *testing.T) {
			result := api.ToGQLClientProvider(test.provider)
			assert.Equal(t, result, test.expected)
		})
	}
}
