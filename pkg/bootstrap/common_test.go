package bootstrap_test

import (
	"testing"

	"github.com/pluralsh/plural/pkg/api"
	"github.com/pluralsh/plural/pkg/bootstrap"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
)

func doNothing(_ []string) error {
	return nil
}

func TestGetStepPath(t *testing.T) {
	tests := []struct {
		name         string
		step         *bootstrap.Step
		defaultPath  string
		expectedPath string
	}{
		{
			name: `step path should be used if it was set`,
			step: &bootstrap.Step{
				Name:       "Test",
				Args:       []string{},
				TargetPath: "/test/path",
				Execute:    doNothing,
			},
			defaultPath:  "/default/path",
			expectedPath: "/test/path",
		},
		{
			name: `step path should be defaulted if not set`,
			step: &bootstrap.Step{
				Name:    "Test",
				Args:    []string{},
				Execute: doNothing,
			},
			defaultPath:  "/default/path",
			expectedPath: "/default/path",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := bootstrap.GetStepPath(test.step, test.defaultPath)
			assert.Equal(t, path, test.expectedPath)
		})
	}
}

func TestFilterSteps(t *testing.T) {
	tests := []struct {
		name          string
		steps         []*bootstrap.Step
		provider      string
		expectedSteps []*bootstrap.Step
	}{
		{
			name: `common steps should not be filtered`,
			steps: []*bootstrap.Step{
				{
					Name:    "Test",
					Execute: doNothing,
				},
			},
			provider: api.ProviderAzure,
			expectedSteps: []*bootstrap.Step{
				{
					Name:    "Test",
					Execute: doNothing,
				},
			},
		},
		{
			name: `steps for different providers should be filtered`,
			steps: []*bootstrap.Step{
				{
					Name:    "Test",
					Execute: doNothing,
				},
				{
					Name:     "Test AWS",
					Execute:  doNothing,
					Provider: api.ProviderAWS,
				},
				{
					Name:     "Test Azure",
					Execute:  doNothing,
					Provider: api.ProviderAzure,
				},
				{
					Name:     "Test GCP",
					Execute:  doNothing,
					Provider: api.ProviderGCP,
				},
			},
			provider: api.ProviderAzure,
			expectedSteps: []*bootstrap.Step{
				{
					Name:    "Test",
					Execute: doNothing,
				},
				{
					Name:     "Test Azure",
					Execute:  doNothing,
					Provider: api.ProviderAzure,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filteredSteps := bootstrap.FilterSteps(test.steps, test.provider)
			assert.Equal(t, len(filteredSteps), len(test.expectedSteps))
			assert.True(t, slices.EqualFunc(filteredSteps, test.expectedSteps,
				func(a *bootstrap.Step, b *bootstrap.Step) bool {
					return a.Name == b.Name && a.Provider == b.Provider
				}))
		})
	}
}
