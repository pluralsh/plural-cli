package bootstrap_test

import (
	"fmt"
	"testing"

	"github.com/pluralsh/plural/pkg/bootstrap"
	"github.com/stretchr/testify/assert"
)

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
				Execute: func(_ []string) error {
					return nil
				},
			},
			defaultPath:  "/default/path",
			expectedPath: "/test/path",
		},
		{
			name: `step path should be defaulted if not set`,
			step: &bootstrap.Step{
				Name: "Test",
				Args: []string{},
				Execute: func(_ []string) error {
					return nil
				},
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

func TestExecuteSteps(t *testing.T) {
	tests := []struct {
		name        string
		steps       []*bootstrap.Step
		expectError bool
	}{
		{
			name: `steps should be executed successfully`,
			steps: []*bootstrap.Step{
				{
					Name:       "Test 1",
					Args:       []string{},
					TargetPath: ".",
					Execute: func(_ []string) error {
						return nil
					},
				},
				{
					Name:       "Test 2",
					Args:       []string{},
					TargetPath: ".",
					Execute: func(_ []string) error {
						return nil
					},
				},
			},
			expectError: false,
		},
		{
			name: `steps should be executed successfully if args are not set`,
			steps: []*bootstrap.Step{
				{
					Name:       "Test",
					TargetPath: ".",
					Execute: func(_ []string) error {
						return nil
					},
				},
			},
			expectError: false,
		},
		{
			name: `steps execution should fail on invalid path`,
			steps: []*bootstrap.Step{
				{
					Name:       "Test",
					TargetPath: "invalid-path",
					Execute: func(_ []string) error {
						return nil
					},
				},
			},
			expectError: true,
		},
		{
			name: `steps execution should fail on execution error`,
			steps: []*bootstrap.Step{
				{
					Name:       "Test",
					TargetPath: ".",
					Execute: func(_ []string) error {
						return fmt.Errorf("error")
					},
				},
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := bootstrap.ExecuteSteps(test.steps)
			if test.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
