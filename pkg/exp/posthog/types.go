package posthog

type Client interface {
	IsFeatureEnabled(FeatureFlagPayload) (bool, error)
}

type FeatureFlagPayload struct {
	Key        string
	DistinctId string
}

type Config struct {
	APIKey   string
	Endpoint string
}
