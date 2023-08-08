package bootstrap_test

import (
	"fmt"
	"testing"

	"github.com/pluralsh/plural/pkg/bootstrap"
	"github.com/stretchr/testify/assert"
)

func TestGetProviderTags(t *testing.T) {
	providers := []string{"aws", "azure", "google"}

	for _, provider := range providers {
		t.Run(fmt.Sprintf("test %s tags", provider), func(t *testing.T) {
			tags := bootstrap.GetProviderTags(provider, "test")
			_, err := bootstrap.GetProviderTagsMap(tags)
			assert.NoError(t, err)
		})
	}
}

func TestGetProviderTagsMap(t *testing.T) {
	tests := []struct {
		name           string
		arguments      []string
		expectedResult map[string]string
		expectError    bool
	}{
		{
			name:           `tags should be returned successfully`,
			arguments:      []string{"test=abc", "qwerty=test"},
			expectedResult: map[string]string{"test": "abc", "qwerty": "test"},
			expectError:    false,
		},
		{
			name:           `tags should be returned successfully if arguments are empty`,
			arguments:      []string{},
			expectedResult: map[string]string{},
			expectError:    false,
		},
		{
			name:           `error should be returned if arguments are in invalid format`,
			arguments:      []string{"invalid-format"},
			expectedResult: nil,
			expectError:    true,
		},
		{
			name:           `error should be returned if arguments are in invalid format`,
			arguments:      []string{"valid=tag", "invalid=format=test"},
			expectedResult: nil,
			expectError:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := bootstrap.GetProviderTagsMap(test.arguments)
			if test.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.expectedResult, result)
		})
	}
}
