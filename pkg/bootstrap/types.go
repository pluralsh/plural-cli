package bootstrap

import "github.com/pluralsh/cluster-api-migration/pkg/api"

// ActionFunc is an action function that is executed as a part of single bootstrap, migrate and destroy step.
type ActionFunc func(arguments []string) error

// ConditionFunc is a condition function that is checks if step should be executed or skipped.
type ConditionFunc func() bool

type HookFunc func()

// Step is a representation of a single step in a process of bootstrap, migrate and destroy.
type Step struct {
	Name       string
	Args       []string
	TargetPath string
	Execute    ActionFunc
	Skip       bool
	SkipFunc   ConditionFunc
	Retries    int
	OnError    HookFunc
	OnAfter    HookFunc
}

// Bootstrap is a representation of existing cluster to be migrated to Cluster API.
// This data is fetched from provider with migrator tool.
// See github.com/pluralsh/cluster-api-migration for more details.
type Bootstrap struct {
	ClusterAPICluster *api.Values `json:"cluster-api-cluster"`
}