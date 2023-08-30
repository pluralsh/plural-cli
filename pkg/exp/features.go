package exp

const (
	EXP_PLURAL_CAPI = FeatureFlag("EXP_PLURAL_CAPI")
)

var (
	providers = []Provider{
		newEnvProvider(),
		newPostHogProvider(),
	}
)

func IsFeatureEnabled(feature FeatureFlag) bool {
	for _, p := range providers {
		if p.IsFeatureEnabled(feature) {
			return true
		}
	}

	return false
}
