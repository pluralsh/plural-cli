package up

import (
	"fmt"
	"strings"
)

func validateClusterName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("cluster name is required")
	}
	if len(name) > 15 {
		return fmt.Errorf("cluster name must be at most 15 characters")
	}
	// Matches utils.ValidateAlphaNumeric used by provider init.
	if len(name) < 2 || name[0] < 'a' || name[0] > 'z' {
		return fmt.Errorf("cluster name must start with a lowercase letter")
	}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			continue
		}
		return fmt.Errorf("cluster name must be lowercase alphanumeric with hyphens")
	}
	return nil
}

// ValidateProviderForm checks required self-hosted provider fields.
func ValidateProviderForm(providerID string, values map[string]string) error {
	for _, field := range ProviderFormFields(providerID) {
		val := strings.TrimSpace(values[field.Key])
		if field.Required && val == "" {
			return fmt.Errorf("%s is required", field.Label)
		}
		if field.Key == "cluster" {
			if err := validateClusterName(val); err != nil {
				return err
			}
		}
	}
	return nil
}
