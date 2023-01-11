package utils_test

import (
	"testing"

	"github.com/pluralsh/plural/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestPatchInterfaceMap(t *testing.T) {
	tests := []struct {
		name                  string
		defaultValues, values map[string]map[string]interface{}
		expectedResult        map[string]map[string]interface{}
		expectError           bool
	}{
		{
			name: `test if map are equal`,
			defaultValues: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13},
			},
			values: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13},
			},
			expectedResult: map[string]map[string]interface{}{},
		},
		{
			name: `test if new element was added`,
			defaultValues: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13},
			},
			values: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13, "c": "d"},
			},
			expectedResult: map[string]map[string]interface{}{
				"test": {"c": "d"},
			},
		},
		{
			name: `test if element was changed`,
			defaultValues: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13},
			},
			values: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13, "c": "test"},
			},
			expectedResult: map[string]map[string]interface{}{
				"test": {"c": "test"},
			},
		},
		{
			name: `test with empty values`,
			defaultValues: map[string]map[string]interface{}{
				"test": {"a": "test", "b": 13},
			},
			values:         map[string]map[string]interface{}{},
			expectedResult: map[string]map[string]interface{}{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := utils.PatchInterfaceMap(test.defaultValues, test.values)
			if test.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.expectedResult, res)
		})
	}
}
