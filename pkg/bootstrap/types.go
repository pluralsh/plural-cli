package bootstrap

import "github.com/pluralsh/cluster-api-migration/pkg/api"

// ActionFunc is an action function that is executed as a part of single bootstrap, migrate and destroy step.
type ActionFunc func(arguments []string) error

type ConditionFunc func() bool

// Step is a representation of a single step in a process of bootstrap, migrate and destroy.
type Step struct {
	Name             string
	Args             []string
	TargetPath       string
	BootstrapCommand bool
	Execute          ActionFunc
	Skip             ConditionFunc
}

// Bootstrap is a representation of existing cluster to be migrated to Cluster API.
// This data is fetched from provider with migrator tool.
// See github.com/pluralsh/cluster-api-migration for more details.
type Bootstrap struct {
	ClusterAPICluster *api.Values `json:"cluster-api-cluster"`
}
