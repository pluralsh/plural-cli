package api

import (
	"strings"
)

func NormalizeProvider(prov string) string {
	provider := strings.ToUpper(prov)
	if provider == "GOOGLE" {
		return "GCP"
	}

	return provider
}
