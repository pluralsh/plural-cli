package api_test

import (
	"testing"

	"github.com/pluralsh/gqlclient"
	"github.com/pluralsh/plural/pkg/api"

	"github.com/stretchr/testify/assert"
)

func TestConstructRecipe(t *testing.T) {
	key := "key"
	tests := []struct {
		name     string
		input    string
		expected gqlclient.RecipeAttributes
	}{
		{
			name: `test ConstructRecipe method`,
			expected: gqlclient.RecipeAttributes{
				OidcSettings: &gqlclient.OidcSettingsAttributes{
					DomainKey: &key,
				},
			},
			input: `
oidcSettings:
  domainKey: "key"
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repositoryAttributes, err := api.ConstructRecipe([]byte(test.input))
			assert.NoError(t, err)
			assert.Equal(t, test.expected, repositoryAttributes)
		})
	}
}
