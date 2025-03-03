package api

import (
	"strings"
)

func NormalizeProvider(p string) string {
	// Compare with ignore case
	if strings.EqualFold(p, ProviderGCPDeprecated) {
		return ProviderGCP
	}

	return p
}

func ToGQLClientProvider(p string) string {
	return strings.ToUpper(NormalizeProvider(p))
}
