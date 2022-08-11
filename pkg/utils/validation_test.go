package utils_test

import (
	"testing"

	"github.com/pluralsh/plural/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestValidateSingleLevelDeep(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		subdomain   string
		expectError bool
	}{
		{
			name:        `test correct domain`,
			domain:      "test.example.com",
			subdomain:   "example.com",
			expectError: false,
		},
		{
			name:        `test too long domain`,
			domain:      "test.test.example.com",
			subdomain:   "example.com",
			expectError: true,
		},
		{
			name:        `test incorrect domain`,
			domain:      "test..example.com",
			subdomain:   "example.com",
			expectError: true,
		},
		{
			name:        `test the same as subdomain`,
			domain:      "example.com",
			subdomain:   "example.com",
			expectError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := utils.ValidateSingleLevelDeep(test.domain, test.subdomain)
			if test.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
