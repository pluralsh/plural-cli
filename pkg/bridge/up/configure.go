package up

import (
	"fmt"
	"strings"

	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/utils"
)

const pluralDNSDomain = "onplural.sh"

// BucketPrefixPrompt matches ProjectManifest.Configure (self-hosted).
const BucketPrefixPrompt = "Enter a unique, memorable string to use for bucket naming, e.g. an abbreviation for your company:"

// PluralSubdomainPrompt matches ProjectManifest.ConfigureNetwork.
const PluralSubdomainPrompt = "Enter subdomain of onplural.sh domain that you want to use:"

// ValidateBucketPrefix mirrors Configure's bucket-name validator.
func ValidateBucketPrefix(val string) error {
	return utils.ValidateRegex(val, "[a-z][0-9\\-a-z]+", "bucket name can only contain alphanumeric characters or hyphens")
}

// PluralDomain builds subdomain.onplural.sh (or returns a full onplural.sh name).
func PluralDomain(subdomain string) string {
	subdomain = strings.TrimSpace(subdomain)
	if strings.HasSuffix(subdomain, pluralDNSDomain) {
		return subdomain
	}
	return subdomain + "." + pluralDNSDomain
}

// ValidatePluralSubdomain checks DNS shape for the Plural DNS subdomain prompt.
func ValidatePluralSubdomain(subdomain string) error {
	return utils.ValidateDns(PluralDomain(subdomain))
}

// RegisterPluralDomain creates the Plural DNS domain (same as ConfigureNetwork).
func RegisterPluralDomain(subdomain string) (string, error) {
	d := PluralDomain(subdomain)
	if err := utils.ValidateDns(d); err != nil {
		return "", err
	}
	client := api.NewClient()
	if err := client.CreateDomain(d); err != nil {
		return "", fmt.Errorf("domain %s is taken or your user doesn't have sufficient permissions to create domains", subdomain)
	}
	return d, nil
}
