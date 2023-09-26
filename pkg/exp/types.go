package exp

type FeatureFlag string

type Provider interface {
	IsFeatureEnabled(feature FeatureFlag) bool
}
