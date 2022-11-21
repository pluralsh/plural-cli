package api

import (
	"strings"

	"github.com/pluralsh/polly/algorithms"
	"github.com/samber/lo"
)

func NormalizeProvider(prov string) string {
	provider := strings.ToUpper(prov)
	if provider == "GOOGLE" {
		return "GCP"
	}

	return provider
}

func FromSlicePtr[T any](s []*T) []T {
	return algorithms.Map(s, lo.FromPtr[T])
}
