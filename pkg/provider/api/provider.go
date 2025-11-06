package api

import (
	"github.com/pluralsh/plural-cli/pkg/provider/permissions"
	"github.com/pluralsh/plural-cli/pkg/provider/preflights"
)

type Provider interface {
	Name() string
	Cluster() string
	Project() string
	Region() string
	Bucket() string
	KubeConfig() error
	KubeContext() string
	CreateBucket() error
	Context() map[string]interface{}
	Preflights() []*preflights.Preflight
	Permissions() (permissions.Checker, error)
	Flush() error
}
