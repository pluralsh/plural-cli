package api

import (
	"strings"

	"github.com/pluralsh/polly/algorithms"
	"github.com/samber/lo"
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

func FromSlicePtr[T any](s []*T) []T {
	return algorithms.Map(s, lo.FromPtr[T])
}
