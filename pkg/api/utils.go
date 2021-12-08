package api

import (
	"strings"
)

func NormalizeProvider(prov string) string {
	provider := strings.ToUpper(prov) 
	if prov == "GOOGLE" {
		return "GCP"
	}

	return provider
}