package api_test

import (
	"testing"

	"github.com/pluralsh/plural/pkg/api"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeProvider(t *testing.T) {
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
			result := api.NormalizeProvider(test.provider)
			assert.Equal(t, result, test.expected)
		})
	}
}
