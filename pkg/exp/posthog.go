package exp

import (
	"github.com/pluralsh/plural/pkg/config"
	"github.com/pluralsh/plural/pkg/exp/posthog"
	"github.com/pluralsh/plural/pkg/utils"
)

// PostHogProvider implements Provider interface
type PostHogProvider struct {
	cache  map[FeatureFlag]bool
	client posthog.Client
	email  string
}

func (this *PostHogProvider) IsFeatureEnabled(feature FeatureFlag) bool {
	if enabled, exists := this.fromCache(feature); exists {
		return enabled
	}

	isEnabled, err := this.client.IsFeatureEnabled(posthog.FeatureFlagPayload{
		Key:        string(feature),
		DistinctId: this.email,
	})

	if err != nil {
		utils.Warn("%s", err)
		return false
	}

	return isEnabled
}

func (this *PostHogProvider) fromCache(feature FeatureFlag) (enabled, exists bool) {
	enabled, exists = this.cache[feature]
	return
}

func (this *PostHogProvider) init() Provider {
	this.client = posthog.New()
	this.email = config.Read().Email
	this.cache = make(map[FeatureFlag]bool, 0)

	return this
}

func newPostHogProvider() Provider {
	return (&PostHogProvider{}).init()
}
