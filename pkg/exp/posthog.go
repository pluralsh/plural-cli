package exp

import (
	"github.com/pluralsh/plural-cli/pkg/config"
	"github.com/pluralsh/plural-cli/pkg/exp/posthog"
)

// PostHogProvider implements Provider interface
type PostHogProvider struct {
	cache  map[FeatureFlag]bool
	client posthog.Client
	email  string
}

func (php *PostHogProvider) IsFeatureEnabled(feature FeatureFlag) bool {
	if enabled, exists := php.fromCache(feature); exists {
		return enabled
	}

	isEnabled, err := php.client.IsFeatureEnabled(posthog.FeatureFlagPayload{
		Key:        string(feature),
		DistinctId: php.email,
	})

	if err != nil {
		// We can cache it for the CLI to avoid retries
		php.cache[feature] = false
		return false
	}

	php.cache[feature] = isEnabled
	return isEnabled
}

func (php *PostHogProvider) fromCache(feature FeatureFlag) (enabled, exists bool) {
	enabled, exists = php.cache[feature]
	return
}

func (php *PostHogProvider) init() Provider {
	php.client = posthog.New()
	php.email = config.Read().Email
	php.cache = make(map[FeatureFlag]bool, 0)

	return php
}
