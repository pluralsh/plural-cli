package exp

import (
	"os"
)

type EnvProvider struct{}

func (this *EnvProvider) IsFeatureEnabled(feature FeatureFlag) bool {
	return os.Getenv(string(feature)) == "true"
}

func newEnvProvider() Provider {
	return &EnvProvider{}
}
